package tunnel

import (
	"sync"

	"guacamole-client-go/pkg/session"
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

type SessionCache struct {
	sync.Mutex
	Sessions map[string]*session.TunnelSession
}

func (g *SessionCache) Add(s *session.TunnelSession) {
	g.Lock()
	defer g.Unlock()
	g.Sessions[s.ID] = s
}

func (g *SessionCache) Pop(sid string) *session.TunnelSession {
	g.Lock()
	defer g.Unlock()
	sess := g.Sessions[sid]
	delete(g.Sessions, sid)
	return sess
}
