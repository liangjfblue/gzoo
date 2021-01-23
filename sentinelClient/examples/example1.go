package main

import (
	"fmt"
	"gzoo/common/sentinelClient"
	"log"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

// setAndGet 测试正常读写
func setAndGet(sc sentinelClient.SentinelClient) {
	conn := sc.GetMasterClient()
	defer conn.Close()

	if _, err := redis.Int(conn.Do("SET", "name", "aaa")); err != nil {
		log.Fatal(fmt.Sprintf("set err:%v", err))
	}

	resp, err := redis.String(conn.Do("GET", "name"))
	if err != nil {
		log.Fatal(fmt.Sprintf("set err:%v", err))
	}
	fmt.Println(resp)
}

// autoSwitchMaster 测试主从切换
func autoSwitchMaster(sc sentinelClient.SentinelClient) {
	var wg1 sync.WaitGroup

	wg1.Add(1)
	go func() {
		defer wg1.Done()

		conn := sc.GetMasterClient()
		defer conn.Close()

		if _, err := redis.Int(conn.Do("SET", "name", "aaa")); err != nil {
			log.Fatal(fmt.Sprintf("set err:%v", err))
		}

		resp, err := redis.String(conn.Do("GET", "name"))
		if err != nil {
			log.Fatal(fmt.Sprintf("get err:%v", err))
		}
		fmt.Println(resp)
	}()

	wg1.Wait()

	// TODO kill掉master, 测试选主
	time.Sleep(time.Second * 10)

	var wg2 sync.WaitGroup

	wg2.Add(1)
	go func() {
		defer wg2.Done()

		conn := sc.GetMasterClient()
		defer conn.Close()

		if _, err := redis.Int(conn.Do("SET", "name", "bbb")); err != nil {
			log.Fatal(fmt.Sprintf("set err:%v", err))
		}

		resp, err := redis.String(conn.Do("GET", "name"))
		if err != nil {
			log.Fatal(fmt.Sprintf("get err:%v", err))
		}
		fmt.Println(resp)
	}()
	wg2.Wait()
}

func recvSwitchMasterCallback(text string) {
	fmt.Println(text)
}

func main() {
	sc := sentinelClient.New()
	if err := sc.Init(
		sentinelClient.SentinelHosts([]string{"127.0.0.1:11001", "127.0.0.1:11002", "127.0.0.1:11003"}),
		sentinelClient.MasterName("test-sentinel"),
		sentinelClient.MaxIdle(16),
		sentinelClient.MaxActive(64),
		sentinelClient.DialConnTimeout(time.Second*5),
		sentinelClient.DialTimeout(time.Second*5),
		sentinelClient.IdleCheckTime(time.Second*5),
		sentinelClient.MonitorStatusDuration(time.Second*5),
		sentinelClient.SwitchMasterCallback(recvSwitchMasterCallback),
	); err != nil {
		log.Fatal(fmt.Sprintf("init sentinel err:%v", err))
	}
	defer sc.Close()

	setAndGet(sc)
	autoSwitchMaster(sc)
}
