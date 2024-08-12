package guacd

import (
	"testing"
)

func TestValidateInstructionString(t *testing.T) {
	tests := []string{
		"1.a,2.bc,3.def,10.helloworld;",
		"4.test,5.test2;",
		"0.;",
		"3.foo;",
		"4.args,13.VERSION_1_3_0,8.hostname,4.port,6.domain,8.username,8.password," +
			"5.width,6.height,3.dpi,15.initial-program,11.color-depth," +
			"13.disable-audio,15.enable-printing,12.printer-name,12.enable-drive,10.drive-name," +
			"10.drive-path,17.create-drive-path,16.disable-download,14.disable-upload," +
			"7.console,13.console-audio,13.server-layout,8.security,11.ignore-cert," +
			"12.disable-auth,10.remote-app,14.remote-app-dir,15.remote-app-args,15.static-channels," +
			"11.client-name,16.enable-wallpaper,14.enable-theming,21.enable-font-smoothing,23.enable-full-window-drag," +
			"26.enable-desktop-composition,22.enable-menu-animations,22.disable-bitmap-caching,25.disable-offscreen-caching,21.disable-glyph-caching,16.preconnection-id,18.preconnection-blob,8.timezone,11.enable-sftp,13.sftp-hostname,13.sftp-host-key,9.sftp-port,13.sftp-username,13.sftp-password,16.sftp-private-key,15.sftp-passphrase,14.sftp-directory,19.sftp-root-directory,26.sftp-server-alive-interval,21.sftp-disable-download,19.sftp-disable-upload,14.recording-path,14.recording-name,24.recording-exclude-output,23.recording-exclude-mouse,22.recording-include-keys,21.create-recording-path,13.resize-method,18.enable-audio-input,9.read-only,16.gateway-hostname,12.gateway-port,14.gateway-domain,16.gateway-username,16.gateway-password,17.load-balance-info," +
			"12.disable-copy,13.disable-paste,15.wol-send-packet,12.wol-mac-addr,18.wol-broadcast-addr,13.wol-wait-time;",
		"5.audio,1.1,31.audio/L16;",
		"5e.audio,1.1,31.audio/L16;",
		";",
	}

	for i := range tests {
		ins, err := ParseInstructionString(tests[i])
		if err != nil {
			t.Log(err)
			continue
		}
		t.Log(ins)

	}

}
