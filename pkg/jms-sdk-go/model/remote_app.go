package model

import "fmt"

type Applet struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type AppletOption struct {
	ID              string                 `json:"id"`
	Applet          Applet                 `json:"applet"`
	Host            Asset                  `json:"host"`
	Account         Account                `json:"account"`
	Gateway         *Gateway               `json:"gateway"`
	Platform        *Platform              `json:"platform"`
	RemoteAppOption RemoteAppCommandOption `json:"remote_app_option"`
	Error           string                 `json:"error"`
}

type RemoteAppCommandOption struct {
	Program string `json:"remoteapplicationprogram:s"`
	Name    string `json:"remoteapplicationname:s"`
	Shell   string `json:"alternate shell:s"`
	CmdLine string `json:"remoteapplicationcmdline:s"`
}

type VirtualAppContainer struct {
	ContainerId string `json:"container_id"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Protocol    string `json:"protocol"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	SFTPPort    int    `json:"sftp_port"`
}

func (v VirtualAppContainer) String() string {
	return fmt.Sprintf("%s://%s:%d",
		v.Protocol, v.Host, v.Port)
}

type VirtualAppOption struct {
	ImageName     string `json:"image_name"`
	ImageProtocol string `json:"image_protocol"`
	ImagePort     int    `json:"image_port"`
	DesktopWidth  int    `json:"desktop_width"`
	DesktopHeight int    `json:"desktop_height"`
}

type VirtualApp struct {
	Name          string `json:"name"`
	ImageName     string `json:"image_name"`
	ImageProtocol string `json:"image_protocol"`
	ImagePort     int    `json:"image_port"`
	Error         string `json:"error"`
}
