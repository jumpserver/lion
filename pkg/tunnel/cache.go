package tunnel

import (
	"sync"
)

type GuaTunnelCache struct {
	sync.Mutex
	Tunnels map[string]*Connection
}

func (g *GuaTunnelCache) Add(t *Connection) {
	g.Lock()
	defer g.Unlock()
	g.Tunnels[t.guacdTunnel.UUID] = t
}

func (g *GuaTunnelCache) Delete(t *Connection) {
	g.Lock()
	defer g.Unlock()
	delete(g.Tunnels, t.guacdTunnel.UUID)
}

func (g *GuaTunnelCache) Get(tid string) *Connection {
	g.Lock()
	defer g.Unlock()
	return g.Tunnels[tid]
}
