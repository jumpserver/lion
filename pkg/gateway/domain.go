package gateway

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"guacamole-client-go/pkg/common"
	"guacamole-client-go/pkg/jms-sdk-go/model"
)

var ErrNoAvailable = errors.New("no available domain")

const (
	miniTimeout = 15 * time.Second
)

type DomainGateway struct {
	Domain  *model.Domain
	DstAddr string // 10.0.0.1:3389

	sshClient       *gossh.Client
	selectedGateway *model.Gateway
	ln              net.Listener

	once sync.Once

	err error
}

func (d *DomainGateway) run() {
	defer d.closeOnce()
	for {
		con, err := d.ln.Accept()
		if err != nil {
			break
		}
		go d.handlerConn(con)
	}
}

func (d *DomainGateway) handlerConn(srcCon net.Conn) {
	defer srcCon.Close()
	dstCon, err := d.sshClient.Dial("tcp", d.DstAddr)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer dstCon.Close()
	go func() {
		_, _ = io.Copy(dstCon, srcCon)
		_ = dstCon.Close()
	}()
	_, _ = io.Copy(srcCon, dstCon)
}

func (d *DomainGateway) Start() (err error) {
	if !d.getAvailableGateway() {
		return ErrNoAvailable
	}
	localIP := common.CurrentLocalIP()
	d.ln, err = net.Listen("tcp", net.JoinHostPort(localIP, "0"))
	if err != nil {
		_ = d.sshClient.Close()
		return err
	}
	go d.run()
	return nil
}

func (d *DomainGateway) GetListenAddr() *net.TCPAddr {
	return d.ln.Addr().(*net.TCPAddr)
}

func (d *DomainGateway) getAvailableGateway() bool {
	for i := range d.Domain.Gateways {
		gateway := d.Domain.Gateways[i]
		if gateway.Protocol == "ssh" {
			auths := make([]gossh.AuthMethod, 0, 3)
			if gateway.Password != "" {
				auths = append(auths, gossh.Password(gateway.Password))
				auths = append(auths, gossh.KeyboardInteractive(func(user, instruction string,
					questions []string, echos []bool) (answers []string, err error) {
					return []string{gateway.Password}, nil
				}))
			}
			if gateway.PrivateKey != "" {
				if signer, err := gossh.ParsePrivateKey([]byte(gateway.PrivateKey)); err == nil {
					auths = append(auths, gossh.PublicKeys(signer))
				}
			}
			sshConfig := gossh.ClientConfig{
				User:            gateway.Username,
				Auth:            auths,
				HostKeyCallback: gossh.InsecureIgnoreHostKey(),
				Timeout:         miniTimeout,
			}
			addr := net.JoinHostPort(gateway.IP, strconv.Itoa(gateway.Port))
			sshClient, err := gossh.Dial("tcp", addr, &sshConfig)
			if err != nil {
				continue
			}
			d.sshClient = sshClient
			d.selectedGateway = &gateway
			return true
		}
	}
	return false
}

func (d *DomainGateway) Stop() {
	d.closeOnce()
}

func (d *DomainGateway) closeOnce() {
	d.once.Do(func() {
		_ = d.ln.Close()
		_ = d.sshClient.Close()
	})
}
