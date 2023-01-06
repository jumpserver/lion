package tunnel

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	ginSessions "github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"lion/pkg/common"
	"lion/pkg/config"
	"lion/pkg/gateway"
	"lion/pkg/guacd"
	"lion/pkg/jms-sdk-go/model"
	"lion/pkg/jms-sdk-go/service"
	"lion/pkg/logger"
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
	SessCache      *SessionCache
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
	sessionId, ok := ctx.GetQuery("SESSION_ID")
	if !ok {
		logger.Error("No session id params")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrBadParams.String()))
		return
	}
	userItem, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		logger.Error("No auth user found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}

	tunnelSession := g.SessCache.Pop(sessionId)
	if tunnelSession == nil {
		logger.Error("No session found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrNoSession.String()))
		return
	}
	defer func() {
		if err2 := tunnelSession.ReleaseAppletAccount(); err2 != nil {
			logger.Errorf("Release account failed: %s", err2)

		}
	}()
	if user := userItem.(*model.User); user.ID != tunnelSession.User.ID {
		logger.Error("No valid auth user found")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAuthUser.String()))
		return
	}

	if err = tunnelSession.ConnectedCallback(); err != nil {
		logger.Errorf("Session connect callback err %v", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrAPIFailed.String()))
		return
	}
	info := g.getClientInfo(ctx)
	conf := tunnelSession.GuaConfiguration()
	// 设置网域网关，替换本地。 兼容云平台同步 配置网域，但网关配置为空的情况
	if (tunnelSession.Domain != nil && len(tunnelSession.Domain.Gateways) != 0) ||
		tunnelSession.Gateway != nil {
		dstAddr := net.JoinHostPort(conf.GetParameter(guacd.Hostname),
			conf.GetParameter(guacd.Port))
		domainGateway := gateway.DomainGateway{
			Domain:  tunnelSession.Domain,
			DstAddr: dstAddr,

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
	guacdAddr := net.JoinHostPort(config.GlobalConfig.GuaHost, config.GlobalConfig.GuaPort)
	tunnel, err = guacd.NewTunnel(guacdAddr, conf, info)
	if err != nil {
		logger.Errorf("Connect tunnel err: %+v", err)
		_ = ws.WriteMessage(websocket.TextMessage, []byte(ErrGuacamoleServer.String()))
		if err = tunnelSession.ConnectedFailedCallback(err); err != nil {
			logger.Errorf("Update session connect status failed %+v", err)
		}
		if err = tunnelSession.DisConnectedCallback(); err != nil {
			logger.Errorf("Session DisConnectedCallback err: %+v", err)
		}
		return
	}
	defer tunnel.Close()
	if err := tunnelSession.ConnectedSuccessCallback(); err != nil {
		logger.Errorf("Update session connect status failed %+v", err)
	}
	conn := Connection{
		Sess:        tunnelSession,
		guacdTunnel: tunnel,
		Service:     g.SessionService,
		ws:          ws,
		done:        make(chan struct{}),
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
	if err = tunnelSession.FinishReplayCallback(); err != nil {
		logger.Errorf("Session Replay upload err: %+v", err)
	}
	logger.Infof("Session[%s] disconnect", sessionId)
}

func (g *GuacamoleTunnelServer) CreateSession(ctx *gin.Context) {
	var jsonData struct {
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.BindJSON(&jsonData); err != nil {
		logger.Errorf("Token session json invalid: %+v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	value, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		logger.Error("No auth user found")
		ctx.JSON(http.StatusBadRequest, ErrorResponse(ErrNoAuthUser))
		return
	}
	user := value.(*model.User)
	connectSession, err := g.SessionService.CreatByToken(ctx, jsonData.Token)
	if err != nil {
		logger.Errorf("Create token session err: %+v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	if user.ID != connectSession.User.ID {
		logger.Errorf("No match connect token user %s but got %s",
			connectSession.User.String(), user.String())
		ctx.JSON(http.StatusBadRequest, ErrorResponse(ErrNoAuthUser))
		return
	}
	g.SessCache.Add(&connectSession)
	ctx.JSON(http.StatusCreated, SuccessResponse(connectSession))
}

func (g *GuacamoleTunnelServer) TokenSession(ctx *gin.Context) {
	var jsonData struct {
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.BindJSON(&jsonData); err != nil {
		logger.Errorf("Token session json invalid: %+v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	connectSession, err := g.SessionService.CreatByToken(ctx, jsonData.Token)
	if err != nil {
		logger.Errorf("Create token session err: %+v", err)
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	ginAuthSession := ginSessions.Default(ctx)
	ginAuthSession.Set(config.GinSessionKey, connectSession.User.ID)
	if err = ginAuthSession.Save(); err != nil {
		logger.Errorf("Save gin session err: %s", err.Error())
		ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}
	g.SessCache.Add(&connectSession)
	ctx.JSON(http.StatusCreated, SuccessResponse(connectSession))
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
		fileLog := model.FTPLog{
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.String(),
			OrgID:      tun.Sess.Asset.OrgID,
			SystemUser: tun.Sess.Account.String(),
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateDownload,
			Path:       filename,
			DateStart:  common.NewNowUTCTime(),
		}
		ctx.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		out := OutStreamResource{
			streamIndex: index,
			mediaType:   "",
			writer:      ctx.Writer,
			ctx:         ctx.Request.Context(),
			done:        make(chan struct{}),
		}
		tun.outputFilter.addOutStream(out)
		if err := out.Wait(); err != nil {
			logger.Errorf("Session[%s] download file %s err: %s", tun, filename, err)
			ctx.JSON(http.StatusBadRequest, ErrorResponse(err))
			g.SessionService.AuditFileOperation(fileLog)
			return
		}
		fileLog.IsSuccess = true
		g.SessionService.AuditFileOperation(fileLog)
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
		fileLog := model.FTPLog{
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.String(),
			OrgID:      tun.Sess.Asset.OrgID,
			SystemUser: tun.Sess.Account.String(),
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateUpload,
			Path:       filename,
			DateStart:  common.NewNowUTCTime(),
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
			_ = fdReader.Close()
			if err := stream.WaitErr(); err != nil {
				logger.Errorf("Session[%s] upload file %s err: %s", tun, filename, err)
				g.SessionService.AuditFileOperation(fileLog)
				continue
			}
			logger.Infof("Session[%s] upload file %s success", tun, filename)
			fileLog.IsSuccess = true
			g.SessionService.AuditFileOperation(fileLog)
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
	conn := MonitorCon{
		guacdTunnel: tunnelCon,
		ws:          ws,
	}
	logger.Infof("User %s start to monitor session %s", user, sessionId)
	_ = conn.Run(ctx.Request.Context())
	g.Cache.RemoveMonitorTunneler(sessionId, tunnelCon)
	logger.Infof("User %s stop to monitor session %s", user, sessionId)
}

func (g *GuacamoleTunnelServer) DeleteSession(ctx *gin.Context) {
	sid := ctx.Param("sid")
	if sess := g.SessCache.Pop(sid); sess != nil {
		logger.Infof("Delete session %s", sid)
		return
	}
	ctx.JSON(http.StatusOK, SuccessResponse(nil))
}
