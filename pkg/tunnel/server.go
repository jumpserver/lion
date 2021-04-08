package tunnel

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/model"
	"guacamole-client-go/pkg/jms-sdk-go/service"
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

	tunnelSession.ConnectedSuccessCallback()
	info := g.getClientInfo(ctx)
	conf := tunnelSession.GuaConfiguration()
	var tunnel *guacd.Tunnel

	guacdAddr := net.JoinHostPort(config.GlobalConfig.GuaHost, config.GlobalConfig.GuaPort)
	tunnel, err = guacd.NewTunnel(guacdAddr, conf, info)
	if err != nil {
		fmt.Printf("%+v\n", err)
		data := guacd.NewInstruction(
			guacd.InstructionServerError, err.Error(), "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		return
	}
	defer tunnel.Close()
	conn := Connection{
		sess:        tunnelSession,
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
		ctx.JSON(http.StatusBadRequest, CreateErrorResponse(err))
		return
	}
	fmt.Println("TokenSession: ", jsonData)
	connectSession, err := g.SessionService.CreatByToken(ctx, jsonData.Token)
	if err != nil {
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

	fmt.Println(tid, index, filename)
	if tun := g.Cache.Get(tid); tun != nil {
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
			return
		}
	}
	fmt.Println("DownloadFile ", filename, " ", index, " finished")
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
				fmt.Println("UploadFile ", filename, " ", index, " WaitErr ", err.Error())
			}
			_ = fdReader.Close()
		}
		return
	}
	ctx.AbortWithStatus(http.StatusNotFound)
}
