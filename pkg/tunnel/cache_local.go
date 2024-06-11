package tunnel

import (
	"sync"

	"lion/pkg/guacd"
	"lion/pkg/logger"
)

func NewLocalTunnelLocalCache() *GuaTunnelLocalCache {
	return &GuaTunnelLocalCache{
		Tunnels: make(map[string]*Connection),
	}
}

type GuaTunnelLocalCache struct {
	sync.Mutex
	Tunnels map[string]*Connection
}

func (g *GuaTunnelLocalCache) Add(t *Connection) {
	g.Lock()
	defer g.Unlock()
	g.Tunnels[t.guacdTunnel.UUID()] = t
}

func (g *GuaTunnelLocalCache) Delete(t *Connection) {
	g.Lock()
	defer g.Unlock()
	delete(g.Tunnels, t.guacdTunnel.UUID())
}

func (g *GuaTunnelLocalCache) Get(tid string) *Connection {
	g.Lock()
	defer g.Unlock()
	return g.Tunnels[tid]
}

func (g *GuaTunnelLocalCache) RangeActiveSessionIds() []string {
	g.Lock()
	ret := make([]string, 0, len(g.Tunnels))
	for i := range g.Tunnels {
		ret = append(ret, g.Tunnels[i].Sess.ID)
	}
	g.Unlock()
	return ret
}

func (g *GuaTunnelLocalCache) RangeActiveUserIds() map[string]struct{} {
	g.Lock()
	ret := make(map[string]struct{})
	for i := range g.Tunnels {
		currentUser := g.Tunnels[i].Sess.User
		ret[currentUser.ID] = struct{}{}
	}
	g.Unlock()
	return ret
}

func (g *GuaTunnelLocalCache) GetBySessionId(sid string) *Connection {
	g.Lock()
	defer g.Unlock()
	for i := range g.Tunnels {
		if sid == g.Tunnels[i].Sess.ID {
			return g.Tunnels[i]
		}
	}
	return nil
}

func (g *GuaTunnelLocalCache) GetMonitorTunnelerBySessionId(sid string) Tunneler {
	if conn := g.GetBySessionId(sid); conn != nil {
		if guacdTunnel, err := conn.CloneMonitorTunnel(); err == nil {
			return guacdTunnel
		} else {
			logger.Error(err)
		}
	}
	return nil
}

func (g *GuaTunnelLocalCache) RemoveMonitorTunneler(sid string, monitorTunnel Tunneler) {
	if conn := g.GetBySessionId(sid); conn != nil {
		if tunnel, ok := monitorTunnel.(*guacd.Tunnel); ok {
			conn.unTraceMonitorTunnel(tunnel)
		}
	}
}
