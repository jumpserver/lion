package tunnel

import (
	"sync"
)

type GuaTunnelStorage struct {
	sync.Mutex
	Tunnels map[string]*Connection
}

func (g *GuaTunnelStorage) Add(t *Connection) {
	g.Lock()
	defer g.Unlock()
	g.Tunnels[t.guacdTunnel.UUID] = t
}

func (g *GuaTunnelStorage) Delete(t *Connection) {
	g.Lock()
	defer g.Unlock()
	delete(g.Tunnels, t.guacdTunnel.UUID)
}

func (g *GuaTunnelStorage) Get(tid string) *Connection {
	g.Lock()
	defer g.Unlock()
	return g.Tunnels[tid]
}
