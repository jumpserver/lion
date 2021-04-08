package gateway

import (
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"guacamole-client-go/pkg/jms-sdk-go/model"
)

var ErrNoAvailable = errors.New("no available domain")

const (
	miniTimeout = 15 * time.Second
)

type DomainGateway struct {
	domain  *model.Domain
	dstIP   string
	dstPort int

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
	dstAddr := net.JoinHostPort(d.dstIP, strconv.Itoa(d.dstPort))
	dstCon, err := d.sshClient.Dial("tcp", dstAddr)
	if err != nil {
		return
	}
	defer dstCon.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(dstCon, srcCon)
		_ = dstCon.Close()
	}()
	_, _ = io.Copy(srcCon, dstCon)
	wg.Wait()

}

func (d *DomainGateway) Start() (addr *net.TCPAddr, err error) {
	if !d.getAvailableGateway() {
		return nil, ErrNoAvailable
	}
	d.ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = d.sshClient.Close()
		return nil, err
	}
	go d.run()
	return d.ln.Addr().(*net.TCPAddr), nil
}

func (d *DomainGateway) getAvailableGateway() bool {
	for i := range d.domain.Gateways {
		gateway := d.domain.Gateways[i]
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
