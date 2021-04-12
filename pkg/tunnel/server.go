package tunnel

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/gateway"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
	"guacamole-client-go/pkg/logger"
	"guacamole-client-go/pkg/session"
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
	Cache          *GuaTunnelCache
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
	sessionId, ok := ctx.GetQuery("SESSION_ID")
	if !ok {
		data := guacd.NewInstruction(
			guacd.InstructionServerError, "no session id", "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		return
	}
	tunnelSession := g.SessCache.Pop(sessionId)
	if tunnelSession == nil {
		data := guacd.NewInstruction(
			guacd.InstructionServerError, "no found session", "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		return
	}

	if err = tunnelSession.ConnectedCallback(); err != nil {
		data := guacd.NewInstruction(
			guacd.InstructionServerError, err.Error(), "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		return
	}
	info := g.getClientInfo(ctx)
	conf := tunnelSession.GuaConfiguration()
	// 设置网域网关，替换本地
	if tunnelSession.Domain != nil {
		dstAddr := net.JoinHostPort(conf.GetParameter(guacd.Hostname),
			conf.GetParameter(guacd.Port))
		domainGateway := gateway.DomainGateway{
			Domain:  tunnelSession.Domain,
			DstAddr: dstAddr,
		}
		if err = domainGateway.Start(); err != nil {
			logger.Errorf("Start domain gateway err: %+v", err)
			data := guacd.NewInstruction(
				guacd.InstructionServerError, err.Error(), "504")
			_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
			return
		}
		defer domainGateway.Stop()
		localAddr := domainGateway.GetListenAddr()
		conf.SetParameter(guacd.Hostname, localAddr.IP.String())
		conf.SetParameter(guacd.Port, strconv.Itoa(localAddr.Port))
		logger.Infof("Start domain gateway %s listen on %s:%d", tunnelSession.Domain.Name,
			localAddr.IP.String(), localAddr.Port)
	}

	var tunnel *guacd.Tunnel
	guacdAddr := net.JoinHostPort(config.GlobalConfig.GuaHost, config.GlobalConfig.GuaPort)
	tunnel, err = guacd.NewTunnel(guacdAddr, conf, info)
	if err != nil {
		logger.Errorf("Connect tunnel err: %+v", err)
		data := guacd.NewInstruction(
			guacd.InstructionServerError, err.Error(), "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		if err = tunnelSession.ConnectedFailedCallback(err); err != nil {
			logger.Errorf("Update session connect status failed %+v", err)
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
		ws:          ws,
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
	g.Cache.Add(&conn)
	err = conn.Run(ctx)
	g.Cache.Delete(&conn)
	if err = tunnelSession.DisConnectedCallback(); err != nil {
		logger.Errorf("Session DisConnectedCallback err: %+v", err)
	}
	if err = tunnelSession.FinishReplayCallback(); err != nil {
		logger.Errorf("Session Replay upload err: %+v", err)
	}
}

func (g *GuacamoleTunnelServer) TokenConnect(ctx *gin.Context) {
	g.Connect(ctx)
}

func (g *GuacamoleTunnelServer) CreateSession(ctx *gin.Context) {
	var jsonData struct {
		TargetId     string `json:"target_id" binding:"required"`
		TargetType   string `json:"type" binding:"required"`
		SystemUserId string `json:"system_user_id" binding:"required"`
	}
	if err := ctx.BindJSON(&jsonData); err != nil {
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	value, ok := ctx.Get(config.GinCtxUserKey)
	if !ok {
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(ErrNoAuthUser))
		return
	}
	user := value.(*model.User)
	connectSession, err := g.SessionService.Creat(ctx, user,
		jsonData.TargetType, jsonData.TargetId, jsonData.SystemUserId)
	if err != nil {
		logger.Errorf("Create session err: %+v", err)
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	g.SessCache.Add(&connectSession)
	ctx.JSON(http.StatusCreated, CreateSuccessResponse(connectSession))
}

func (g *GuacamoleTunnelServer) TokenSession(ctx *gin.Context) {
	var jsonData struct {
		Token string `json:"token" binding:"required"`
	}
	if err := ctx.BindJSON(&jsonData); err != nil {
		logger.Errorf("Token session json invalid: %+v", err)
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	connectSession, err := g.SessionService.CreatByToken(ctx, jsonData.Token)
	if err != nil {
		logger.Errorf("Create token session err: %+v", err)
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	g.SessCache.Add(&connectSession)
	ctx.JSON(http.StatusCreated, CreateSuccessResponse(connectSession))
}

func (g *GuacamoleTunnelServer) DownloadFile(ctx *gin.Context) {
	tid := ctx.Param("tid")
	index := ctx.Param("index")
	filename := ctx.Param("filename")

	if tun := g.Cache.Get(tid); tun != nil {
		fileLog := model.FTPLog{
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.Hostname,
			OrgID:      tun.Sess.Asset.OrgID,
			SystemUser: tun.Sess.SystemUser.Name,
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateDownload,
			Path:       filename,
			DataStart:  common.NewNowUTCTime(),
		}
		ctx.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		out := OutStreamResource{
			streamIndex: index,
			mediaType:   "",
			writer:      ctx.Writer,
			done:        make(chan struct{}),
		}
		tun.outputFilter.addOutStream(out)
		if err := out.Wait(); err != nil {
			ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
			_ = g.SessionService.AuditFileOperation(fileLog)
			return
		}
		fileLog.IsSuccess = true
		_ = g.SessionService.AuditFileOperation(fileLog)
		logger.Info("DownloadFile ", filename, " ", index, " finished")
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
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	if tun := g.Cache.Get(tid); tun != nil {
		fileLog := model.FTPLog{
			User:       tun.Sess.User.String(),
			Hostname:   tun.Sess.Asset.Hostname,
			OrgID:      tun.Sess.Asset.OrgID,
			SystemUser: tun.Sess.SystemUser.Name,
			RemoteAddr: ctx.ClientIP(),
			Operate:    model.OperateUpload,
			Path:       filename,
			DataStart:  common.NewNowUTCTime(),
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
				logger.Error("UploadFile ", filename, " ", index, " WaitErr ", err.Error())
				_ = g.SessionService.AuditFileOperation(fileLog)
				continue
			}
			fileLog.IsSuccess = true
			_ = g.SessionService.AuditFileOperation(fileLog)
		}
		return
	}
	ctx.AbortWithStatus(http.StatusNotFound)
}
