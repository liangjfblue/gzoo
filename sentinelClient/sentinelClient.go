package sentinelClient

import (
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

type SentinelClient interface {
	// 初始化
	Init(...Option) error
	// 关闭
	Close()
	// master连接池
	GetMasterClient() redis.Conn
	// slave连接池
	GetSlaverClient() redis.Conn
}

type Option func(*Options)
type SwitchMasterHook func(string)

// redisInfo redis实例信息
type redisInfo struct {
	host       string       // redis host
	mutex      sync.RWMutex // 锁
	conn       redis.Conn   // redis连接
	status     int32        // redis实例状态
	poolMutex  sync.RWMutex // 锁
	poolClient *redis.Pool  // 连接池
}

// sentinelClient sentinel实例
type sentinelClient struct {
	options      Options          // 参数
	stop         chan struct{}    // 关闭标记
	pubSubConn   redis.PubSubConn // 订阅连接
	pubSubStatus int32            // 订阅连接状态
	pubSubMutex  sync.Mutex       // 锁
	master       redisInfo        // redis主
	slaver       redisInfo        // redis从
}

const (
	connectNormal = 0
	connectError  = -100
)

var (
	errOptions           = errors.New("sentinel error options")
	errSwitchQuicklyHost = errors.New("can not get quickly host")
	errGetInfoBySentinel = errors.New("can not get info by sentinel")
)

func New() SentinelClient {
	return &sentinelClient{}
}

// Init 初始化
func (s *sentinelClient) Init(opts ...Option) (err error) {
	// 1.初始化配置
	s.options = defaultOptions
	for _, o := range opts {
		o(&s.options)
	}
	if err = s.checkOptions(); err != nil {
		return err
	}
	s.stop = make(chan struct{}, 1)

	// 2.选出最优sentinel host
	sentinelHost, err := s.switchQuicklyHost(s.options.sentinelHosts)
	if err != nil {
		return err
	}

	// 3.连接sentinel, 获取master信息
	if err = s.connectRedis(sentinelHost); err != nil {
		return err
	}

	// 4.初始化master连接池
	if err = s.initRedisPool(sentinelHost, master, false); err != nil {
		return err
	}

	// 5.初始化slave连接池
	if err = s.initRedisPool(sentinelHost, slave, false); err != nil {
		return err
	}

	// 6.sentinel订阅监听主从切换
	go s.subSentinelEvent()

	// 7.定时检测pubSub和slaveConn的连接
	go s.monitorRedisStatusLoop()

	return
}

// Close 关闭sentinelClient
func (s *sentinelClient) Close() {
	s.stop <- struct{}{}
}

// GetMasterClient 从master连接池获取连接
func (s *sentinelClient) GetMasterClient() redis.Conn {
	s.master.poolMutex.RLock()
	defer s.master.poolMutex.RUnlock()

	return s.master.poolClient.Get()
}

// GetSlaverClient 从slave连接池获取连接
func (s *sentinelClient) GetSlaverClient() redis.Conn {
	s.slaver.poolMutex.RLock()
	defer s.slaver.poolMutex.RUnlock()

	return s.slaver.poolClient.Get()
}

// checkOptions 检查参数
func (s *sentinelClient) checkOptions() error {
	if len(s.options.sentinelHosts) == 0 || len(s.options.masterName) == 0 {
		return errOptions
	}

	if s.options.maxActive <= 8 {
		s.options.maxActive = 8
	}

	if s.options.maxIdle <= 16 {
		s.options.maxActive = 16
	}

	return nil
}

// switchQuicklyHost 选出ping返回最快host
func (s *sentinelClient) switchQuicklyHost(hosts []string) (string, error) {
	if len(hosts) == 0 {
		return "", errSwitchQuicklyHost
	}

	if len(hosts) == 1 {
		return hosts[1], nil
	}

	var indexChan = make(chan int, len(hosts))
	defer close(indexChan)

	for i, host := range hosts {
		go func(i int, host string) {
			client, err := redis.Dial("tcp", host, redis.DialConnectTimeout(s.options.dialConnTimeout))
			if err != nil {
				indexChan <- connectError
				return
			}
			defer client.Close()

			_, err = redis.String(client.Do("ping"))
			if err != nil {
				indexChan <- connectError
			} else {
				indexChan <- i
			}
		}(i, host)
	}

	for index := range indexChan {
		if index != connectError {
			return hosts[index], nil
		}
	}

	return "", errSwitchQuicklyHost
}

// connectRedis 连接redis
func (s *sentinelClient) connectRedis(host string) error {
	conn, err := redis.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()

	s.pubSubConn = redis.PubSubConn{Conn: conn}

	err = s.pubSubConn.Subscribe("+switch-master")

	log.Printf("sentinel subscribe +switch-master, host:%s, err:%v\n", host, err)
	return err
}

// subSentinelEvent sentinel订阅主从切换事件
func (s *sentinelClient) subSentinelEvent() {
	for {
		msg := s.pubSubConn.Receive()
		switch msg.(type) {
		case redis.Message:
			m := msg.(redis.Message)
			s.switchMaster(m.Channel, string(m.Data))
		case redis.Pong:
			log.Printf("%s[master:%s] sentinel monitor msg:%+v\n", s.options.masterName, s.getMasterHost(), msg)
		case error:
			// 重连sentinel
			s.setPubSubStatus(connectError)
		}
	}
}

// monitorRedisStatusLoop 监控sentinel/slave状态
func (s *sentinelClient) monitorRedisStatusLoop() {
	ticker := time.NewTicker(s.options.monitorStatusDuration)
	for {
		select {
		case <-ticker.C:
			// 主从
			switch s.getPubSubStatus() {
			case connectNormal:
				if pong, err := redis.String(s.pubSubConn.Conn.Do("ping")); err != nil {
					s.setPubSubStatus(connectError)
					log.Printf("%s[master:%s] sentinel monitor pong:%v, err:%v\n", s.options.masterName, s.getMasterHost(), pong, err)
				} else {
					log.Printf("%s[master:%s] sentinel monitor pong:%v\n", s.options.masterName, s.getMasterHost(), pong)
				}
			case connectError:
				// 重连sentinel, 开启新监控
				sentinelHost, err := s.switchQuicklyHost(s.options.sentinelHosts)
				if err != nil {
					log.Printf("switch quickly host, err:%v\n", err)
				} else {
					if err := s.connectRedis(sentinelHost); err != nil {
						log.Printf("connect sentinel, err:%v\n", err)
					}
					s.setPubSubStatus(connectNormal)
					s.subSentinelEvent()
				}
			}

			// slave
			switch s.getSlaveStatus() {
			case connectNormal:
				if pong, err := redis.String(s.slaver.conn.Do("ping")); err != nil {
					s.setSlaveStatus(connectError)
					log.Printf("%s[slave:%s] monitor err:%v", s.options.masterName, s.getSlaverHost(), err)
				} else {
					log.Printf("%s[slave:%s] monitor pong:%s", s.options.masterName, s.getSlaverHost(), pong)
				}
			case connectError:
				sentinelHost, err := s.switchQuicklyHost(s.options.sentinelHosts)
				if err != nil {
					log.Printf("switch quickly host, err:%v\n", err)
				} else {
					if err = s.initRedisPool(sentinelHost, slave, false); err != nil {
						log.Printf("switch slave error, cur slave is:%s[%s], err:%+v\n", s.options.masterName, s.getSlaverHost(), err)
					} else {
						s.setSlaveStatus(connectNormal)
						log.Printf("switch slave ok, new slave is:%s[%s]\n", s.options.masterName, s.getSlaverHost())
					}
				}
			}
		}
	}
}

// switchMaster 切换master
func (s *sentinelClient) switchMaster(channel, data string) {
	// 只关注主从切换
	if !strings.HasPrefix(channel, "+switch-master") {
		return
	}

	info := strings.Split(data, " ")
	if len(info) != 5 {
		log.Printf("switch master info err, data:%s\n", data)
		return
	}

	if !strings.EqualFold(info[0], s.options.masterName) {
		log.Printf("switch master the same, cur master:%s, switch data:%+v", s.options.masterName, data)
		return
	}

	masterHost := net.JoinHostPort(info[3], info[4])
	oldMasterHost := s.getMasterHost()

	s.setMasterHost(masterHost)

	s.master.poolMutex.Lock()
	if s.master.poolClient != nil {
		s.master.poolClient.Close()
	}
	s.master.poolClient = s.createRedisPool(masterHost)
	s.master.poolMutex.Unlock()

	// 主从切换回调
	if s.options.switchMasterHook != nil {
		text := fmt.Sprintf("%s(old) --> %s(now)", oldMasterHost, masterHost)
		s.options.switchMasterHook(text)
	}
}
