package common

import (
	"net"
)

func CurrentLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	defer func() {
		if err == nil {
			_ = conn.Close()
		}
	}()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
