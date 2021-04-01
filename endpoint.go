package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/config"
	"guacamole-client-go/pkg/guacd"
	"guacamole-client-go/pkg/jms-sdk-go/service"
)

type GuacamoleTunnelService struct {
	tunnels    map[string]*TunnelConn
	jmsService *service.JMService
}

func (g *GuacamoleTunnelService) createClientInfo(ctx *gin.Context) guacd.ClientInformation {
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

func (g *GuacamoleTunnelService) Connect(ctx *gin.Context) {
	ws, err := upGrader.Upgrade(ctx.Writer, ctx.Request, ctx.Writer.Header())
	if err != nil {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	sessionId, ok := ctx.GetQuery("SESSION_ID")
	if !ok {

		return
	}

	var tunnel *guacd.Tunnel
	//rdpConfig := GetRDPConfiguration(ctx)
	rdpConfig := GetVNCConfiguration(ctx)
	//var rdpConfig guacd.Configuration
	guacdAddr := net.JoinHostPort(config.GlobalConfig.GuaHost, config.GlobalConfig.GuaPort)
	fmt.Println(guacdAddr, sessionId)
	info := g.createClientInfo(ctx)

	tunnel, err = guacd.NewTunnel(guacdAddr, rdpConfig, info)
	if err != nil {
		fmt.Printf("%v\n", err)
		data := guacd.NewInstruction(
			guacd.InstructionServerError, err.Error(), "504")
		_ = ws.WriteMessage(websocket.TextMessage, []byte(data.String()))
		return
	}
	defer tunnel.Close()
	conn := TunnelConn{
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
	g.tunnels[conn.guacdTunnel.UUID] = &conn
	err = conn.Run(ctx)

}

func (g *GuacamoleTunnelService) CreateSession(ctx *gin.Context) {
	Id := common.UUID()
	//data := make(map[string]string)
	ctx.JSON(http.StatusCreated, Id)
}

func (g *GuacamoleTunnelService) download(ctx *gin.Context) {
	tid := ctx.Param("tid")
	index := ctx.Param("index")
	filename := ctx.Param("filename")

	fmt.Println(tid, index, filename)
	if tun, ok := g.tunnels[tid]; ok {
		ctx.Writer.Header().Set("content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		out := OutStreamResource{
			streamIndex: index,
			mediaType:   "",
			writer:      ctx.Writer,
			done:        make(chan struct{}),
		}
		tun.outputFilter.addOutStream(out)
		out.Wait()
	}
	fmt.Println("download ", filename, " ", index, " finished")
}

func (g *GuacamoleTunnelService) upload(ctx *gin.Context) {
	tid := ctx.Param("tid")
	index := ctx.Param("index")
	filename := ctx.Param("filename")
	fmt.Println(tid, index, filename)
	form, err := ctx.MultipartForm()

	if err != nil {
		fmt.Println(err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if tun, ok := g.tunnels[tid]; ok {
		files := form.File["file"]
		for _, file := range files {
			fmt.Println(file.Filename)
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
		}
	}
	fmt.Println("upload ", filename, " ", index, " finished")
}
