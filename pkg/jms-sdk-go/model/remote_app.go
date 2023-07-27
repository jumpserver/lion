package model

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
	RemoteAppOption RemoteAppCommandOption `json:"remote_app_option"`
}

type RemoteAppCommandOption struct {
	Program string `json:"remoteapplicationprogram:s"`
	Name    string `json:"remoteapplicationname:s"`
	Shell   string `json:"alternate shell:s"`
	CmdLine string `json:"remoteapplicationcmdline:s"`
}
