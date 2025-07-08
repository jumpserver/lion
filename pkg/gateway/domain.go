package gateway

import (
	"errors"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	gossh "golang.org/x/crypto/ssh"

	"lion/pkg/logger"

	"github.com/jumpserver-dev/sdk-go/common"
	"github.com/jumpserver-dev/sdk-go/model"
)

var ErrNoAvailable = errors.New("no available domain")

const (
	miniTimeout = 15 * time.Second
)

type DomainGateway struct {
	DstAddr string // 10.0.0.1:3389

	sshClient       *gossh.Client
	SelectedGateway *model.Gateway

	ln net.Listener

	once sync.Once
}

func (d *DomainGateway) run() {
	defer d.closeOnce()
	for {
		con, err := d.ln.Accept()
		if err != nil {
			break
		}
		logger.Infof("Accept new conn by gateway %s ", d.SelectedGateway.Name)
		go d.handlerConn(con)
	}
	logger.Infof("Stop proxy by gateway %s", d.SelectedGateway.Name)
}

func (d *DomainGateway) handlerConn(srcCon net.Conn) {
	defer srcCon.Close()
	dstCon, err := d.sshClient.Dial("tcp", d.DstAddr)
	if err != nil {
		logger.Errorf("Failed gateway dial %s: %s ",
			d.DstAddr, err.Error())
		return
	}
	defer dstCon.Close()
	go func() {
		_, _ = io.Copy(dstCon, srcCon)
		_ = dstCon.Close()
	}()
	_, _ = io.Copy(srcCon, dstCon)
	logger.Infof("Gateway end proxy %s", d.DstAddr)
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
	if d.SelectedGateway != nil {
		sshClient, err := d.createGatewaySSHClient(d.SelectedGateway)
		if err != nil {
			logger.Errorf("Dial select gateway %s err: %s ", d.SelectedGateway.Name, err)
			return false
		}
		logger.Debugf("Dial select gateway %s success", d.SelectedGateway.Name)
		d.sshClient = sshClient
		return true
	}
	return false
}
func (d *DomainGateway) createGatewaySSHClient(gateway *model.Gateway) (*gossh.Client, error) {
	auths := make([]gossh.AuthMethod, 0, 3)
	loginAccount := gateway.Account
	if loginAccount.IsSSHKey() {
		if signer, err1 := gossh.ParsePrivateKey([]byte(loginAccount.Secret)); err1 == nil {
			auths = append(auths, gossh.PublicKeys(signer))
		} else {
			logger.Errorf("Domain gateway Parse private key error: %s", err1)
		}
	} else {
		auths = append(auths, gossh.Password(loginAccount.Secret))
		auths = append(auths, gossh.KeyboardInteractive(func(user, instruction string,
			questions []string, echos []bool) (answers []string, err error) {
			return []string{loginAccount.Secret}, nil
		}))
	}
	sshConfig := gossh.ClientConfig{
		User:            loginAccount.Username,
		Auth:            auths,
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         miniTimeout,
	}
	port := gateway.Protocols.GetProtocolPort("ssh")
	addr := net.JoinHostPort(gateway.Address, strconv.Itoa(port))
	return gossh.Dial("tcp", addr, &sshConfig)
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
