package sentinelClient

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

var (
	defaultOptions = Options{
		maxIdle:               8,
		maxActive:             16,
		dialConnTimeout:       3 * time.Second,
		dialTimeout:           3 * time.Second,
		idleCheckTime:         3 * time.Second,
		monitorStatusDuration: 3 * time.Second,
		switchMasterHook:      nil,
	}
)

type Options struct {
	sentinelHosts         []string           // sentinel host列表
	masterName            string             // master-name
	maxIdle               int                // 连接池 最大空闲连接
	maxActive             int                // 连接池 最大活跃连接
	redisOptions          []redis.DialOption // redis参数
	dialConnTimeout       time.Duration      // 建立连接超时
	dialTimeout           time.Duration      // 读写超时
	idleCheckTime         time.Duration      // 空闲检查时间间隔
	monitorStatusDuration time.Duration      // 监控sentinel/slave状态时间间隔
	switchMasterHook      SwitchMasterHook   // 发生主从切换时的钩子
}

func SentinelHosts(sentinelHosts []string) Option {
	return func(o *Options) {
		o.sentinelHosts = sentinelHosts
	}
}

func MasterName(masterName string) Option {
	return func(o *Options) {
		o.masterName = masterName
	}
}

func MaxIdle(maxIdle int) Option {
	return func(o *Options) {
		o.maxIdle = maxIdle
	}
}

func MaxActive(maxActive int) Option {
	return func(o *Options) {
		o.maxActive = maxActive
	}
}

func RedisOptions(redisOptions []redis.DialOption) Option {
	return func(o *Options) {
		o.redisOptions = redisOptions
	}
}

func DialConnTimeout(dialConnTimeout time.Duration) Option {
	return func(o *Options) {
		o.dialConnTimeout = dialConnTimeout
	}
}

func DialTimeout(dialTimeout time.Duration) Option {
	return func(o *Options) {
		o.dialTimeout = dialTimeout
	}
}

func IdleCheckTime(idleCheckTime time.Duration) Option {
	return func(o *Options) {
		o.idleCheckTime = idleCheckTime
	}
}

func MonitorStatusDuration(monitorStatusDuration time.Duration) Option {
	return func(o *Options) {
		o.monitorStatusDuration = monitorStatusDuration
	}
}

func SwitchMasterCallback(switchMasterCallback SwitchMasterHook) Option {
	return func(o *Options) {
		o.switchMasterHook = switchMasterCallback
	}
}
