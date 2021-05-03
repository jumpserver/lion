package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"

	"lion/pkg/common"
	"lion/pkg/guacd"
	"lion/pkg/logger"
)

const (
	eventsChannel = "JUMPSERVER:LION:EVENTS:CHANNEL"

	resultsChannel = "JUMPSERVER:LION:EVENTS:RESULT"

	sessionsChannelPrefix = "JUMPSERVER:LION:SESSIONS"
)

type Config struct {
	// Addr of a single redis server instance.
	// Defaults to "127.0.0.1:6379".
	Addr string

	Password string

	DBIndex int
}

func NewGuaTunnelRedisCache(conf Config) *GuaTunnelRedisCache {
	if conf.Addr == "" {
		conf.Addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DBIndex,
	})
	if _, err := rdb.Ping(context.TODO()).Result(); err != nil {
		logger.Fatalf("Redis ping err: %s", err)
	}

	cache := GuaTunnelRedisCache{
		ID:                       common.UUID(),
		rdb:                      rdb,
		requestChan:              make(chan *subscribeRequest),
		responseChan:             make(chan chan *subscribeResponse),
		reqCancelChan:            make(chan *subscribeRequest),
		localProxyExitSignalChan: make(chan string, 100),
		GuaTunnelLocalCache:      NewLocalTunnelLocalCache(),
	}
	go cache.run()
	return &cache
}

type GuaTunnelRedisCache struct {
	*GuaTunnelLocalCache

	ID  string
	rdb *redis.Client

	requestChan   chan *subscribeRequest
	responseChan  chan chan *subscribeResponse
	reqCancelChan chan *subscribeRequest

	localProxyExitSignalChan chan string
}

func (r *GuaTunnelRedisCache) GetMonitorTunnelerBySessionId(sid string) Tunneler {
	tunneler := r.GuaTunnelLocalCache.GetMonitorTunnelerBySessionId(sid)
	if tunneler != nil {
		return tunneler
	}
	return r.requestRemoteTunnelerBySessionId(sid)
}

func (r *GuaTunnelRedisCache) requestRemoteTunnelerBySessionId(sid string) Tunneler {
	/*
		1. 发布请求
		2. 收到Tunneler结果
	*/
	req := r.createEventRequest(sid, channelEventJoin)
	res, err := r.sendRequest(&req)
	if err != nil {
		logger.Error(err)
		return nil
	}
	return res.conn
}

func (r *GuaTunnelRedisCache) sendRequest(req *subscribeRequest) (*subscribeResponse, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelFunc()
	r.requestChan <- req
	resultChan := <-r.responseChan
	select {
	case <-ctx.Done():
		select {
		case r.reqCancelChan <- req:

		case res := <-resultChan:
			return res, res.err
		}
	case res := <-resultChan:
		return res, res.err
	}
	return nil, fmt.Errorf("Redis cache send request event %s time out ", req.Event)
}

func (r *GuaTunnelRedisCache) createEventRequest(sid, event string) subscribeRequest {
	reqId := r.uniqueReqId(sid)
	return subscribeRequest{
		ReqId:     reqId,
		SessionId: sid,
		Event:     event,
		Prefix:    reqId,
		Channel:   eventsChannel,
	}
}

func (r *GuaTunnelRedisCache) createResultRequest(reqId, roomId, event string) subscribeRequest {
	return subscribeRequest{
		ReqId:     reqId,
		SessionId: roomId,
		Event:     event,
		Prefix:    reqId,
		Channel:   resultsChannel,
	}
}

/*
(确保每次都是唯一的)
prefix: sessionsChannelPrefix:uuid:reqId:sessionId

*/

func (r *GuaTunnelRedisCache) uniqueReqId(sid string) string {
	return fmt.Sprintf("%s:%s:%s:%s",
		sessionsChannelPrefix,
		common.UUID(),
		r.ID,
		sid)
}

func (r *GuaTunnelRedisCache) publishRequest(req *subscribeRequest) error {
	body, _ := json.Marshal(req)
	return r.publishCommand(req.Channel, body)
}

func (r *GuaTunnelRedisCache) publishCommand(channel string, p []byte) (err error) {
	_, err = r.rdb.Publish(context.TODO(), channel, p).Result()
	return
}

func (r *GuaTunnelRedisCache) proxyTunnel(tunnelProxy *RedisGuacProxy) {
	defer func() {
		_ = tunnelProxy.Close()
		r.GuaTunnelLocalCache.RemoveMonitorTunneler(tunnelProxy.sessionId, tunnelProxy.tunnel)
	}()
	logger.Infof("Redis guacd proxy %s tunnel start", tunnelProxy.reqId)
	for {
		ins, err := tunnelProxy.ReadInstruction()
		if err != nil {
			logger.Errorf("Redis guacd proxy %s tunnel exit", tunnelProxy.reqId)
			return
		}
		if err = r.publishCommand(tunnelProxy.writeChannelName, []byte(ins.String())); err != nil {
			logger.Errorf("Redis guacd proxy %s pubSub message err: %s", tunnelProxy.reqId, err)
		}
	}
}

func (r *GuaTunnelRedisCache) run() {
	innerPubSub := r.rdb.Subscribe(context.TODO(), eventsChannel, resultsChannel)
	subscribeEventsMsgCh := innerPubSub.Channel()
	requestsMap := make(map[string]chan *subscribeResponse)
	proxyConnMap := make(map[string]*RedisGuacProxy)
	for {
		select {
		case redisMsg := <-subscribeEventsMsgCh:
			var req subscribeRequest
			if err := json.Unmarshal([]byte(redisMsg.Payload), &req); err != nil {
				logger.Errorf("Redis cache unmarshal request msg err: %s", err)
				continue
			}
			logger.Infof("Redis channel %s recv request event %s",
				redisMsg.Channel, req.Event)

			switch redisMsg.Channel {
			case eventsChannel:
				if _, ok := requestsMap[req.ReqId]; ok {
					logger.Infof("Redis cache ignore self request %s", req.ReqId)
					continue
				}
				// 创建result channel的req
				switch req.Event {
				case channelEventJoin:
					successReq := r.createResultRequest(req.ReqId, req.SessionId,
						channelEventJoinSuccess)
					if conn := r.GuaTunnelLocalCache.GetBySessionId(req.SessionId); conn != nil {
						guacdTunnel, err := conn.CloneMonitorTunnel()
						if err != nil {
							logger.Errorf("Redis cache create monitor tunneler for request %s: %s",
								req.ReqId, err)
							continue
						}
						err = r.publishRequest(&successReq)
						if err != nil {
							_ = guacdTunnel.Close()
							logger.Errorf("Redis cache reply request %s join event err %s", req.ReqId, err)
							continue
						}
						logger.Infof("Redis cache reply request %s join event", req.ReqId)
						writeChannel := fmt.Sprintf("%s.read", req.Prefix)
						readChannel := fmt.Sprintf("%s.write", req.Prefix)
						pubSub := r.rdb.Subscribe(context.TODO(), readChannel)
						proxyConn := RedisGuacProxy{
							reqId:            req.ReqId,
							sessionId:        req.SessionId,
							readChannelName:  readChannel,
							writeChannelName: writeChannel,
							pubSub:           pubSub,
							cache:            r,
							done:             make(chan struct{}),
							tunnel:           guacdTunnel,
						}
						proxyConnMap[req.ReqId] = &proxyConn
						go proxyConn.run()
						go r.proxyTunnel(&proxyConn)
					}

				case channelEventExit:
					if proxyConn, ok := proxyConnMap[req.ReqId]; ok {
						logger.Infof("Redis cache reply %s exit event", req.ReqId)
						delete(proxyConnMap, req.ReqId)
						successReq := r.createResultRequest(req.ReqId, req.SessionId,
							channelEventExitSuccess)
						err := r.publishRequest(&successReq)
						if err != nil {
							logger.Errorf("Redis cache reply request %s exit event err %s", req.ReqId, err)
							continue
						}
						_ = proxyConn.Close()
					}
				}

			case resultsChannel:
				responseChan, ok := requestsMap[req.ReqId]
				if !ok {
					logger.Debugf("Redis cache ignore not self result request %s", req.ReqId)
					continue
				}
				logger.Infof("Redis cache request %s receive result event %s", req.ReqId, req.Event)
				// 请求结束，移除缓存, 返回请求的结果
				delete(requestsMap, req.ReqId)
				switch req.Event {
				case channelEventJoinSuccess:
					var res subscribeResponse
					res.Req = &req
					writeChannel := fmt.Sprintf("%s.write", req.Prefix)
					readChannel := fmt.Sprintf("%s.read", req.Prefix)
					pubSub := r.rdb.Subscribe(context.TODO(), readChannel)
					conn := RedisConn{
						reqId:            req.ReqId,
						sessionId:        req.SessionId,
						readChannelName:  readChannel,
						writeChannelName: writeChannel,
						instructionChan:  make(chan guacd.Instruction, 100),
						cache:            r,
						pubSub:           pubSub,
						done:             make(chan struct{}),
					}
					res.conn = &conn
					go conn.run()
					responseChan <- &res
				case channelEventExitSuccess:
					var res subscribeResponse
					res.Req = &req
					responseChan <- &res
				}
			default:
				continue
			}
		case req := <-r.requestChan:
			logger.Debugf("Redis cache publish request %s event %s", req.ReqId, req.Event)
			responseChan := make(chan *subscribeResponse, 1)
			r.responseChan <- responseChan
			if err := r.publishRequest(req); err != nil {
				logger.Errorf("Redis cache publish channel request err: %s", err)
				delete(requestsMap, req.ReqId)
				responseChan <- &subscribeResponse{
					Req:  req,
					err:  err,
					conn: nil,
				}
				continue
			}
			logger.Infof("Redis cache publish request %s event %s success", req.ReqId, req.Event)
			requestsMap[req.ReqId] = responseChan

		case req := <-r.reqCancelChan:
			delete(requestsMap, req.ReqId)
			logger.Debugf("Redis cache cancel request: %s", req.ReqId)

		case reqId := <-r.localProxyExitSignalChan:
			if _, ok := proxyConnMap[reqId]; ok {
				logger.Infof("Redis cache recv proxy con %s exit signal", reqId)
				delete(proxyConnMap, reqId)
			}
		}
	}

}

type RedisConn struct {
	reqId     string
	sessionId string

	readChannelName  string
	writeChannelName string
	instructionChan  chan guacd.Instruction
	cache            *GuaTunnelRedisCache
	once             sync.Once
	pubSub           *redis.PubSub

	done chan struct{}
}

func (r *RedisConn) run() {
	logger.Infof("Redis Conn %s pubSub run", r.reqId)
	messageChan := r.pubSub.Channel()
	defer close(r.instructionChan)
	detectTicker := time.NewTicker(time.Minute)
	defer detectTicker.Stop()
	activeTime := time.Now()
	for {
		select {
		case detectTime := <-detectTicker.C:
			if detectTime.After(activeTime.Add(5 * time.Minute)) {
				logger.Errorf("Redis Conn %s time out after 5 minute and exit.", r.reqId)
				return
			}
			continue
		case <-r.done:
			return
		case msg, ok := <-messageChan:
			if !ok {
				logger.Infof("Redis Conn %s pubSub exit", r.reqId)
				return
			}
			switch msg.Channel {
			case r.readChannelName:
				if ret, err := guacd.ParseInstructionString(msg.Payload); err == nil {
					select {
					case <-r.done:
						return
					case r.instructionChan <- ret:
					}
				} else {
					logger.Errorf("Redis Conn %s parse instruction err: %+v", r.reqId, err)
				}
			}
		}
		activeTime = time.Now()
	}
}

func (r *RedisConn) WriteAndFlush(p []byte) (int, error) {
	if err := r.cache.publishCommand(r.writeChannelName, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (r *RedisConn) ReadInstruction() (guacd.Instruction, error) {
	if instruction, ok := <-r.instructionChan; ok {
		return instruction, nil
	}
	return guacd.Instruction{}, io.EOF
}

func (r *RedisConn) Close() error {
	var err error
	r.once.Do(func() {
		logger.Debugf("Redis conn %s close", r.reqId)
		err = r.pubSub.Close()
		_, err := r.cache.sendRequest(&subscribeRequest{
			ReqId:     r.reqId,
			SessionId: r.sessionId,
			Event:     channelEventExit,
			Prefix:    r.reqId,
			Channel:   eventsChannel,
		})
		if err != nil {
			logger.Errorf("Redis conn %s send exit event err: %s", r.reqId, err)
		}

	})
	return err
}

const (
	channelEventJoin        = "Join"
	channelEventExit        = "Exit"
	channelEventJoinSuccess = "JoinSuccess"
	channelEventExitSuccess = "ExitSuccess"
)

type subscribeRequest struct {
	ReqId     string `json:"req_id"`
	SessionId string `json:"session_id"`
	Event     string `json:"event"`
	Prefix    string `json:"prefix"`
	Channel   string `json:"-"`
}

type subscribeResponse struct {
	Req  *subscribeRequest
	err  error
	conn *RedisConn
}

type RedisGuacProxy struct {
	reqId     string
	sessionId string

	readChannelName  string
	writeChannelName string
	pubSub           *redis.PubSub

	cache *GuaTunnelRedisCache

	done chan struct{}

	tunnel *guacd.Tunnel

	once sync.Once
}

func (r *RedisGuacProxy) run() {
	logger.Infof("Redis guacd proxy %s pubSub run", r.reqId)
	redisMsgChan := r.pubSub.Channel()
	defer func() {
		r.tunnel.Close()
		r.cache.localProxyExitSignalChan <- r.reqId
	}()
	for {
		select {
		case redisMsg, ok := <-redisMsgChan:
			if !ok {
				logger.Infof("Redis guacd proxy %s pubSub exit", r.reqId)
				return
			}
			_, _ = r.tunnel.WriteAndFlush([]byte(redisMsg.Payload))
		case <-r.done:
			return
		}
	}
}

func (r *RedisGuacProxy) ReadInstruction() (guacd.Instruction, error) {
	return r.tunnel.ReadInstruction()
}

func (r *RedisGuacProxy) Close() error {
	var err error
	r.once.Do(func() {
		err = r.pubSub.Close()
		close(r.done)
		logger.Infof("Redis guacd proxy %s close", r.reqId)
	})
	return err
}
