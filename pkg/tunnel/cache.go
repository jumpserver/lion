package tunnel

import (
	"sync"

	"lion/pkg/guacd"
	"lion/pkg/session"
)

type Tunneler interface {
	UUID() string
	WriteAndFlush(p []byte) (int, error)
	ReadInstruction() (guacd.Instruction, error)
	Close() error
}

type GuaTunnelCache interface {
	Add(*Connection)
	Delete(*Connection)
	Get(string) *Connection
	RangeActiveSessionIds() []string
	RangeActiveUserIds() map[string]struct{}
	GetBySessionId(sid string) *Connection
	GetMonitorTunnelerBySessionId(sid string) Tunneler
	RemoveMonitorTunneler(sid string, monitorTunnel Tunneler)
}

var (
	_ GuaTunnelCache = (*GuaTunnelLocalCache)(nil)
	_ GuaTunnelCache = (*GuaTunnelRedisCache)(nil)
)

type GuaTunnelCacheManager struct {
	GuaTunnelCache
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

func (g *SessionCache) Get(sid string) *session.TunnelSession {
	g.Lock()
	defer g.Unlock()
	return g.Sessions[sid]
}

func (g *SessionCache) Pop(sid string) *session.TunnelSession {
	g.Lock()
	defer g.Unlock()
	sess := g.Sessions[sid]
	delete(g.Sessions, sid)
	return sess
}
