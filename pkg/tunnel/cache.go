package tunnel

import (
	"sync"

	"lion/pkg/session"
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

func (g *GuaTunnelCache) Range() []string {
	g.Lock()
	ret := make([]string, 0, len(g.Tunnels))
	for i := range g.Tunnels {
		ret = append(ret, g.Tunnels[i].Sess.ID)
	}
	g.Unlock()
	return ret
}

func (g *GuaTunnelCache) GetBySessionId(sid string) *Connection {
	g.Lock()
	defer g.Unlock()
	for i := range g.Tunnels {
		if sid == g.Tunnels[i].Sess.ID {
			return g.Tunnels[i]
		}
	}
	return nil
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
