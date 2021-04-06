package session

import (
	"github.com/spf13/viper"

	"guacamole-client-go/pkg/guacd"
)

type DisplayParameter struct {
	Key          string
	DefaultValue string
}

var (
	colorDepth               = DisplayParameter{Key: guacd.RDPColorDepth, DefaultValue: "24"}
	dpi                      = DisplayParameter{Key: guacd.RDPDpi, DefaultValue: ""}
	disableAudio             = DisplayParameter{Key: guacd.RDPDisableAudio}
	enableWallpaper          = DisplayParameter{Key: guacd.RDPEnableWallpaper, DefaultValue: ""}
	enableTheming            = DisplayParameter{Key: guacd.RDPEnableTheming, DefaultValue: ""}
	enableFontSmoothing      = DisplayParameter{Key: guacd.RDPEnableFontSmoothing, DefaultValue: ""}
	enableFullWindowDrag     = DisplayParameter{Key: guacd.RDPEnableFullWindowDrag, DefaultValue: ""}
	enableDesktopComposition = DisplayParameter{Key: guacd.RDPEnableDesktopComposition, DefaultValue: ""}
	enableMenuAnimations     = DisplayParameter{Key: guacd.RDPEnableMenuAnimations, DefaultValue: ""}
	disableBitmapCaching     = DisplayParameter{Key: guacd.RDPDisableBitmapCaching, DefaultValue: ""}
	disableOffscreenCaching  = DisplayParameter{Key: guacd.RDPDisableOffscreenCaching, DefaultValue: ""}
)

type DisPlay struct {
	data map[string]DisplayParameter
}

func (d DisPlay) GetDisplayParams() map[string]string {
	res := make(map[string]string)
	for envKey, displayParam := range d.data {
		res[displayParam.Key] = displayParam.DefaultValue
		if value := viper.GetString(envKey); value != "" {
			res[displayParam.Key] = value
		}
	}
	return res
}

var RDPDisplay = DisPlay{data: map[string]DisplayParameter{
	"JUMPSERVER_COLOR_DEPTH":                colorDepth,
	"JUMPSERVER_DPI":                        dpi,
	"JUMPSERVER_DISABLE_AUDIO":              disableAudio,
	"JUMPSERVER_ENABLE_WALLPAPER":           enableWallpaper,
	"JUMPSERVER_ENABLE_THEMING":             enableTheming,
	"JUMPSERVER_ENABLE_FONT_SMOOTHING":      enableFontSmoothing,
	"JUMPSERVER_ENABLE_FULL_WINDOW_DRAG":    enableFullWindowDrag,
	"JUMPSERVER_ENABLE_DESKTOP_COMPOSITION": enableDesktopComposition,
	"JUMPSERVER_ENABLE_MENU_ANIMATIONS":     enableMenuAnimations,
	"JUMPSERVER_DISABLE_BITMAP_CACHING":     disableBitmapCaching,
	"JUMPSERVER_DISABLE_OFFSCREEN_CACHING":  disableOffscreenCaching,
}}

var VNCDisplay = DisPlay{data: map[string]DisplayParameter{
	"JUMPSERVER_COLOR_DEPTH": colorDepth,
}}

var RDPBuiltIn = map[string]string{
	guacd.RDPDisableGlyphCaching: "true",
}
