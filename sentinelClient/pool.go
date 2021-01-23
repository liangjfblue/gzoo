package sentinelClient

import (
	"log"
	"net"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

const (
	master = iota + 1
	slave
)

func (s *sentinelClient) initRedisPool(sentinelHost string, switchRole int, isClose bool) error {
	switch switchRole {
	case master:
		return s.initMasterRedisPool(sentinelHost, isClose)
	case slave:
		return s.initSlaveRedisPool(sentinelHost, isClose)
	default:
		log.Fatal("init redis role not right, switchRole:", switchRole)
	}
	return nil
}

func (s *sentinelClient) initMasterRedisPool(sentinelHost string, isClosed bool) error {
	conn, err := redis.Dial("tcp", sentinelHost)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := redis.Strings(conn.Do("SENTINEL", "get-master-addr-by-name", s.options.masterName))
	if err != nil {
		return err
	}

	if len(resp) != 2 {
		return errGetInfoBySentinel
	}

	host := net.JoinHostPort(resp[0], resp[1])

	s.master.poolMutex.Lock()
	if s.master.poolClient != nil {
		s.master.poolClient.Close()
	}
	s.master.poolClient = s.createRedisPool(host)
	s.master.poolMutex.Unlock()

	s.setMasterHost(host)

	return err
}

func (s *sentinelClient) initSlaveRedisPool(sentinelHost string, isClosed bool) error {
	conn, err := redis.Dial("tcp", sentinelHost)
	if err != nil {
		return err
	}
	defer conn.Close()

	resp, err := redis.Values(conn.Do("SENTINEL", "slaves", s.options.masterName))
	if err != nil {
		return err
	}

	var slaveHosts []string
	for _, slave := range resp {
		slaveM, err := redis.StringMap(slave, nil)
		if err != nil {
			continue
		}

		slaveHost := slaveM["name"]
		if len(slaveHost) <= 0 {
			log.Println("this slave no name info, slaveM:", slaveM)
			continue
		}

		if !strings.EqualFold(slaveHost, s.getMasterHost()) {
			slaveHosts = append(slaveHosts, slaveHost)
		}
	}

	var quicklySlave string
	if len(slaveHosts) > 0 {
		quicklySlave, err = s.switchQuicklyHost(slaveHosts)
		if err != nil {
			return err
		}
	}

	// slave都挂了, 使用master
	if len(quicklySlave) <= 0 {
		quicklySlave = s.getMasterHost()
	}

	if len(quicklySlave) <= 0 {
		return errSwitchQuicklyHost
	}

	conn, err = redis.Dial(
		"tcp",
		quicklySlave,
		redis.DialConnectTimeout(s.options.dialConnTimeout),
		redis.DialReadTimeout(s.options.dialTimeout),
		redis.DialWriteTimeout(s.options.dialTimeout),
	)
	if err != nil {
		return err
	}

	s.slaver.poolMutex.Lock()
	if s.slaver.poolClient != nil {
		s.slaver.poolClient.Close()
	}
	s.slaver.poolClient = s.createRedisPool(quicklySlave)
	s.slaver.poolMutex.Unlock()

	if isClosed {
		s.slaver.conn.Close()
	}

	s.slaver.conn = conn
	s.setSlaverHost(quicklySlave)
	return err
}

func (s *sentinelClient) createRedisPool(host string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     s.options.maxIdle,
		MaxActive:   s.options.maxActive,
		IdleTimeout: 60 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < s.options.idleCheckTime {
				return nil
			}
			_, err := c.Do("PING")
			if err != nil {
				log.Printf("t:%+v err:%+v\n", t, err)
			}
			return err
		},
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host, s.options.redisOptions...)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
}
