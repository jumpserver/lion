package tunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"lion/pkg/common"
	"lion/pkg/config"
	"lion/pkg/gateway"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
	"lion/pkg/proxy"
	"lion/pkg/session"
)

const (
	defaultBufferSize = 1024
)

var upGrader = websocket.Upgrader{
	ReadBufferSize:  defaultBufferSize,
	WriteBufferSize: defaultBufferSize,
	Subprotocols:    []string{"guacamole"},
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type GuacamoleTunnelServer struct {
	JmsService     *service.JMService
	Cache          *GuaTunnelCacheManager
	SessionService *session.Server
}

func (g *GuacamoleTunnelServer) getClientInfo(ctx *gin.Context) guacd.ClientInformation {
	info := guacd.NewClientInformation()
	if supportImages, ok := ctx.GetQueryArray("GUAC_IMAGE"); ok {
		info.ImageMimetypes = supportImages
	}
	if supportAudios, ok := ctx.GetQueryArray("GUAC_AUDIO"); ok {
		info.AudioMimetypes = supportAudios
	}
	if supportVideos, ok := ctx.GetQueryArray("GUAC_VIDEO"); ok {
		info.VideoMimetypes = supportVideos
	}
	if width, ok := ctx.GetQuery("GUAC_WIDTH"); ok {
		if widthInt, err := strconv.Atoi(width); err == nil && widthInt > 0 {
			info.OptimalScreenWidth = widthInt
		}
	}
	if height, ok := ctx.GetQuery("GUAC_HEIGHT"); ok {
		if heightInt, err := strconv.Atoi(height); err == nil && heightInt > 0 {
			info.OptimalScreenHeight = heightInt
		}
	}

	if dpi, ok := ctx.GetQuery("GUAC_DPI"); ok {
		if dpiInt, err := strconv.Atoi(dpi); err == nil && dpiInt > 0 {
			info.OptimalResolution = dpiInt
		}
	}
	if keyboardLayout, ok := ctx.GetQuery("GUAC_KEYBOARD"); ok {
		info.KeyboardLayout = keyboardLayout
	}
	return info
}

func (g *GuacamoleTunnelServer) Connect(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, ctx.Writer.Header())
	if err != nil {
		logger.Errorf("Websocket Upgrade err: %+v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer ws.Close()

	tokenId, ok := ctx.GetQuery("TOKEN_ID")
	if !ok {
		logger.Error("No TOKEN id params")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}
	tunnelSession, err := g.SessionService.CreatByToken(ctx, tokenId)
	if err != nil {
		logger.Errorf("Create token session err: %+v", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAPIFailed.String()))
		return
	}
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		logger.Error("No auth user found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}

	defer func() {
		if err2 := tunnelSession.ReleaseAppletAccount(); err2 != nil {
			logger.Errorf("Release account failed: %s", err2)

		}
	}()
	user := userItem.(*model.User)
	if user.ID != tunnelSession.User.ID {
		logger.Error("No valid auth user found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}
	sessionId := tunnelSession.ID
	logger.Infof("User %s start to connect session %s", user, sessionId)
	if err = tunnelSession.ConnectedCallback(); err != nil {
		logger.Errorf("Session connect callback err %v", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAPIFailed.String()))
		return
	}
	p, _ := json.Marshal(tunnelSession)
	ins := NewJmsEventInstruction("session", string(p))
	if err = ws.WriteMessage(websocket.TextMessage, []byte(ins.String())); err != nil {
		logger.Errorf("Write session message err: %+v", err)
		return
	}
	if tunnelSession.AuthInfo.Ticket != nil {
		if err2 := g.JmsService.CreateSessionTicketRelation(tunnelSession.ID,
			tunnelSession.AuthInfo.Ticket.ID); err2 != nil {
			logger.Errorf("Create Session %s Ticket %s relation err: %s", tunnelSession.ID,
				tunnelSession.AuthInfo.Ticket.ID, err2)
		}
	}
	info := g.getClientInfo(ctx)
	opts := tunnelSession.AuthInfo.ConnectOptions
	resolution := strings.ToLower(opts.Resolution)
	switch resolution {
	case "":
	case "auto":
	default:
		logger.Infof("Session[%s] Connect options resolution: %s",
			sessionId, resolution)
		resolutions := strings.Split(resolution, "x")
		if len(resolutions) == 2 {
			width := resolutions[0]
			height := resolutions[1]
			if widthInt, err1 := strconv.Atoi(width); err1 == nil && widthInt > 0 {
				info.OptimalScreenWidth = widthInt
			}
			if heightInt, err1 := strconv.Atoi(height); err1 == nil && heightInt > 0 {
				info.OptimalScreenHeight = heightInt
			}
		}
	}

	conf := tunnelSession.GuaConfiguration()
	for argName, argValue := range info.ExtraConfig() {
		conf.SetParameter(argName, argValue)
	}
	if tunnelSession.Gateway != nil {
		dstAddr := net.JoinHostPort(conf.GetParameter(guacd.Hostname),
			conf.GetParameter(guacd.Port))
		domainGateway := gateway.DomainGateway{
			DstAddr:         dstAddr,
			SelectedGateway: tunnelSession.Gateway,
		}
		if err = domainGateway.Start(); err != nil {
			logger.Errorf("Start domain gateway err: %+v", err)
			_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrGatewayFailed.String()))
			if err = tunnelSession.ConnectedFailedCallback(err); err != nil {
				logger.Errorf("Update session connect status failed %+v", err)
			}
			if err = tunnelSession.DisConnectedCallback(); err != nil {
				logger.Errorf("Session DisConnectedCallback err: %+v", err)
			}
			return
		}
		defer domainGateway.Stop()
		localAddr := domainGateway.GetListenAddr()
		conf.SetParameter(guacd.Hostname, localAddr.IP.String())
		conf.SetParameter(guacd.Port, strconv.Itoa(localAddr.Port))
		logger.Infof("Start domain gateway %s listen on %s:%d", domainGateway.SelectedGateway.Name,
			localAddr.IP.String(), localAddr.Port)
	}

	var tunnel *guacd.Tunnel
	guacdAddr := config.GlobalConfig.SelectGuacdAddr()

	tunnel, err = guacd.NewTunnel(guacdAddr, conf, info)
	if err != nil {
		logger.Errorf("Connect tunnel err: %+v", err)
		msg := fmt.Sprintf("Connect guacd server %s failed", guacdAddr)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrGuacamoleServer.String()))
		if err = tunnelSession.ConnectedFailedCallback(err); err != nil {
			logger.Errorf("Update session connect status failed %+v", err)
		}
		if err = tunnelSession.DisConnectedCallback(); err != nil {
			logger.Errorf("Session DisConnectedCallback err: %+v", err)
		}
		reason := model.SessionLifecycleLog{Reason: msg}
		g.RecordLifecycleLog(sessionId, model.AssetConnectFinished, reason)
		return
	}
	defer tunnel.Close()
	g.RecordLifecycleLog(sessionId, model.AssetConnectSuccess, model.EmptyLifecycleLog)
	if err1 := tunnelSession.ConnectedSuccessCallback(); err1 != nil {
		logger.Errorf("Update session connect status failed %+v", err1)
	}
	logger.Infof("Session[%s] use resolution (%d*%d)",
		sessionId, info.OptimalScreenWidth, info.OptimalScreenHeight)
	meta := MetaMessage{
		ShareId:    user.ID,
		UserId:     user.ID,
		User:       user.String(),
		Created:    time.Now().UTC().String(),
		RemoteAddr: ctx.ClientIP(),
		Primary:    true,
		Writable:   true,
	}
	conn := Connection{
		guacdAddr:   guacdAddr,
		Sess:        &tunnelSession,
		guacdTunnel: tunnel,
		Service:     g.SessionService,
		ws:          ws,
		done:        make(chan struct{}),
		Cache:       g.Cache,
		meta:        &meta,

		currentOnlineUsers: make(map[string]MetaMessage),
	}
	outFilter := OutputStreamInterceptingFilter{
		acknowledgeBlobs: true,
		tunnel:           &conn,
		streams:          map[string]OutStreamResource{},
	}
	inputFilter := InputStreamInterceptingFilter{
		tunnel:  &conn,
		streams: map[string]*InputStreamResource{},
	}
	conn.outputFilter = &outFilter
	conn.inputFilter = &inputFilter
	logger.Infof("Session[%s] connect success", sessionId)
	g.Cache.Add(&conn)
	_ = conn.Run(ctx)
	g.Cache.Delete(&conn)
	if err = tunnelSession.DisConnectedCallback(); err != nil {
		logger.Errorf("Session DisConnectedCallback err: %+v", err)
	}
	if err = tunnelSession.FinishReplayCallback(info); err != nil {
		logger.Errorf("Session Replay upload err: %+v", err)
	}
	logger.Infof("Session[%s] disconnect", sessionId)
}

func (g *GuacamoleTunnelServer) RecordLifecycleLog(sid string, event model.LifecycleEvent,
	logObj model.SessionLifecycleLog) {
	if err := g.JmsService.RecordSessionLifecycleLog(sid, event, logObj); err != nil {
		logger.Errorf("Record session %s lifecycle %s log err: %s", sid, event, err)
	}
}

func (g *GuacamoleTunnelServer) DownloadFile(ctx *gin.Context) {
	tid := ctx.Param("tid")
	index := ctx.Param("index")
	filename := ctx.Param("filename")
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		logger.Error("Download file but not user authorized")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userItem.(*model.User)
	if tun := g.Cache.Get(tid); tun != nil && tun.Sess.User.ID == user.ID {
		recorder := proxy.GetFTPFileRecorder(g.JmsService)
		fileLog := model.FTPLog{
			ID:         common.UUID(),
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.String(),
			OrgID:      tun.Sess.Asset.OrgID,
			Account:    tun.Sess.Account.String(),
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateDownload,
			Path:       filename,
			DateStart:  common.NewNowUTCTime(),
			Session:    tun.Sess.ID,
		}
		ctx.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		out := OutStreamResource{
			streamIndex: index,
			mediaType:   "",
			writer:      ctx.Writer,
			ctx:         ctx.Request.Context(),
			done:        make(chan struct{}),
			ftpLog:      &fileLog,
			recorder:    recorder,
		}
		tun.outputFilter.addOutStream(out)
		if err := out.Wait(); err != nil {
			logger.Errorf("Session[%s] download file %s err: %s", tun, filename, err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			g.SessionService.AuditFileOperation(fileLog)
			recorder.RemoveFtpLog(fileLog.ID)
			return
		}
		fileLog.IsSuccess = true
		g.SessionService.AuditFileOperation(fileLog)
		recorder.FinishFTPFile(fileLog.ID)
		logger.Infof("Session[%s] download file %s success", tun, filename)
		return
	}
	ctx.AbortWithStatus(http.StatusNotFound)
}

func (g *GuacamoleTunnelServer) UploadFile(ctx *gin.Context) {
	tid := ctx.Param("tid")
	index := ctx.Param("index")
	filename := ctx.Param("filename")
	form, err := ctx.MultipartForm()
	if err != nil {
		logger.Errorf("Upload file err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		logger.Error("Upload file but not user authorized")
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	user := userItem.(*model.User)
	if tun := g.Cache.Get(tid); tun != nil && tun.Sess.User.ID == user.ID {
		logger.Infof("User %s upload file %s", user, filename)
		recorder := proxy.GetFTPFileRecorder(g.JmsService)
		fileLog := model.FTPLog{
			ID:         common.UUID(),
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.String(),
			OrgID:      tun.Sess.Asset.OrgID,
			Account:    tun.Sess.Account.String(),
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateUpload,
			Path:       filename,
			DateStart:  common.NewNowUTCTime(),
			Session:    tun.Sess.ID,
		}
		files := form.File["file"]
		for _, file := range files {
			fdReader, err := file.Open()
			if err != nil {
				return
			}
			stream := InputStreamResource{
				streamIndex: index,
				reader:      fdReader,
				done:        make(chan struct{}),
			}
			tun.inputFilter.addInputStream(&stream)
			stream.Wait()
			if err := stream.WaitErr(); err != nil {
				logger.Errorf("Session[%s] upload file %s err: %s", tun, filename, err)
				g.SessionService.AuditFileOperation(fileLog)
				continue
			}
			logger.Infof("Session[%s] upload file %s success", tun, filename)
			fileLog.IsSuccess = true
			g.SessionService.AuditFileOperation(fileLog)
			_, _ = fdReader.(io.Seeker).Seek(0, io.SeekStart)
			if err1 := recorder.Record(&fileLog, fdReader); err1 != nil {
				logger.Errorf("Record file %s err: %s", filename, err1)
			}
			_ = fdReader.Close()
		}
		return
	}
	logger.Infof("No session tunnel found")
	ctx.AbortWithStatus(http.StatusNotFound)
}

func (g *GuacamoleTunnelServer) Monitor(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, ctx.Writer.Header())
	if err != nil {
		logger.Errorf("Websocket Upgrade err: %+v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer ws.Close()
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}
	user := userItem.(*model.User)
	if user.ID == "" {
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}
	sessionId, ok := ctx.GetQuery("SESSION_ID")
	if !ok {
		logger.Error("No session param found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}

	result, err := g.JmsService.ValidateJoinSessionPermission(user.ID, sessionId)
	if err != nil {
		logger.Errorf("Validate join session err: %s", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAPIFailed.String()))
		return
	}
	if !result.Ok {
		logger.Errorf("Validate join session failed : %s", result.Msg)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrPermission.String()))
		return
	}

	tunnelCon := g.Cache.GetMonitorTunnelerBySessionId(sessionId)
	if tunnelCon == nil {
		logger.Error("No session tunnel found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrNoSession.String()))
		return
	}
	defer tunnelCon.Close()

	ints := guacd.NewInstruction(INTERNALDATAOPCODE, tunnelCon.UUID())
	_ = ws.WriteMessage(websocket.TextMessage, []byte(ints.String()))
	conn := MonitorCon{
		Id:          sessionId,
		guacdTunnel: tunnelCon,
		ws:          ws,
		Service:     g,
		User:        user,
	}
	logger.Infof("User %s start to monitor session %s", user, sessionId)
	logObj := model.SessionLifecycleLog{User: user.String()}
	g.RecordLifecycleLog(sessionId, model.AdminJoinMonitor, logObj)
	defer func() {
		g.RecordLifecycleLog(sessionId, model.AdminExitMonitor, logObj)
	}()
	_ = conn.Run(ctx.Request.Context())
	g.Cache.RemoveMonitorTunneler(sessionId, tunnelCon)
	logger.Infof("User %s stop to monitor session %s", user, sessionId)
}

func (g *GuacamoleTunnelServer) CreateShare(ctx *gin.Context) {
	var params struct {
		SessionId   string   `json:"session_id"`
		ExpiredTime int      `json:"expired_time"`
		Users       []string `json:"users"`
		ActionPerm  string   `json:"action_perm"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		logger.Errorf("Bind share params err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	shareReq := model.SharingSessionRequest{
		SessionID:  params.SessionId,
		ExpireTime: params.ExpiredTime,
		Users:      params.Users,
		ActionPerm: params.ActionPerm,
	}
	logger.Debugf("Create share room %v", shareReq)
	if resp, err := g.JmsService.CreateShareRoom(shareReq); err != nil {
		logger.Errorf("Create share room err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
	} else {
		ctx.JSON(http.StatusOK, resp)
		g.RecordLifecycleLog(params.SessionId, model.CreateShareLink,
			model.EmptyLifecycleLog)
	}
}

func (g *GuacamoleTunnelServer) GetShare(ctx *gin.Context) {
	var params struct {
		Code string `json:"code"`
	}
	if err := ctx.BindJSON(&params); err != nil {
		logger.Errorf("Bind share params err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	shareId := ctx.Param("id")
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		err1 := fmt.Errorf("not auth user")
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err1))
		return
	}
	user := userItem.(*model.User)
	data := model.SharePostData{
		ShareId:    shareId,
		Code:       params.Code,
		UserId:     user.ID,
		RemoteAddr: ctx.ClientIP(),
	}
	recordRet, err := g.JmsService.JoinShareRoom(data)
	if err != nil {
		logger.Errorf("Validate join session err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	if recordRet.Err != nil {
		logger.Errorf("Join share room err: %s", recordRet.Err)
		ctx.JSON(http.StatusBadRequest, recordRet.Err)
		return
	}
	ctx.JSON(http.StatusOK, recordRet)
}

func (g *GuacamoleTunnelServer) Share(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, ctx.Writer.Header())
	if err != nil {
		logger.Errorf("Websocket Upgrade err: %+v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer ws.Close()
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}
	user := userItem.(*model.User)
	if user.ID == "" {
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}
	shareId, ok := ctx.GetQuery("SHARE_ID")
	if !ok {
		logger.Error("No share id params")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}
	recordId, ok := ctx.GetQuery("RECORD_ID")
	if !ok {
		logger.Error("No record id params")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}
	sessionId, ok := ctx.GetQuery("SESSION_ID")
	if !ok {
		logger.Error("No SESSION_ID found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}
	writable := false
	writePem, _ := ctx.GetQuery("Writable")
	writable = strings.EqualFold(writePem, "true")
	fmt.Println(writable)

	logger.Debugf("User %s start to share session %s", user, sessionId)
	tunnelCon := g.Cache.GetMonitorTunnelerBySessionId(sessionId)
	if tunnelCon == nil {
		logger.Error("No session tunnel found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrNoSession.String()))
		return
	}
	defer tunnelCon.Close()
	g.RecordLifecycleLog(sessionId, model.UserJoinSession, model.EmptyLifecycleLog)
	ints := guacd.NewInstruction(INTERNALDATAOPCODE, tunnelCon.UUID())
	_ = ws.WriteMessage(websocket.TextMessage, []byte(ints.String()))
	meta := MetaMessage{
		ShareId:    shareId,
		SessionId:  sessionId,
		UserId:     user.ID,
		User:       user.String(),
		Created:    time.Now().UTC().String(),
		RemoteAddr: ctx.ClientIP(),
		Primary:    false,
		Writable:   writable,
	}
	conn := MonitorCon{
		Id:          sessionId,
		guacdTunnel: tunnelCon,
		ws:          ws,
		Service:     g,
		User:        user,
		Meta:        &meta,
	}
	logger.Infof("User %s start to share session %s", user, sessionId)
	_ = conn.Run(ctx.Request.Context())
	g.Cache.RemoveMonitorTunneler(sessionId, tunnelCon)
	logger.Infof("User %s stop to share session %s", user, sessionId)
	if err1 := g.JmsService.FinishShareRoom(recordId); err1 != nil {
		logger.Errorf("Finish share room err: %s", err1)
	}
	g.RecordLifecycleLog(sessionId, model.UserLeaveSession, model.EmptyLifecycleLog)
}

func (g *GuacamoleTunnelServer) DeleteShare(ctx *gin.Context) {
	var params MetaMessage
	if err := ctx.BindJSON(&params); err != nil {
		logger.Errorf("Bind delete share params err: %s", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		ctx.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "not auth user"})
		return
	}
	user := userItem.(*model.User)
	if user.ID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"ok": false, "msg": "not auth user"})
		return
	}
	var removeData struct {
		User string      `json:"user"`
		Meta MetaMessage `json:"meta"`
	}
	removeData.User = user.String()
	removeData.Meta = params
	jsonData, _ := json.Marshal(removeData)
	logger.Infof("User %s remove share session %s", user, params.SessionId)
	g.Cache.BroadcastSessionEvent(params.SessionId, &Event{
		Type: ShareRemoveUser,
		Data: jsonData,
	})
	ctx.JSON(http.StatusOK, gin.H{"ok": true})
}
