package sentinelClient

import "sync/atomic"

func (s *sentinelClient) getPubSubStatus() int32 {
	return atomic.LoadInt32(&s.pubSubStatus)
}

func (s *sentinelClient) setPubSubStatus(status int32) {
	atomic.StoreInt32(&s.pubSubStatus, status)
}

func (s *sentinelClient) setMasterHost(host string) {
	s.master.mutex.Lock()
	defer s.master.mutex.Unlock()
	s.master.host = host
}

func (s *sentinelClient) getMasterHost() string {
	s.master.mutex.RLock()
	defer s.master.mutex.RUnlock()
	return s.master.host
}

func (s *sentinelClient) setSlaverHost(host string) {
	s.slaver.mutex.Lock()
	defer s.slaver.mutex.Unlock()
	s.slaver.host = host
}

func (s *sentinelClient) getSlaverHost() string {
	s.slaver.mutex.RLock()
	defer s.slaver.mutex.RUnlock()
	return s.slaver.host
}

func (s *sentinelClient) setSlaveStatus(status int32) {
	atomic.StoreInt32(&s.slaver.status, status)
}

func (s *sentinelClient) getSlaveStatus() int32 {
	return atomic.LoadInt32(&s.slaver.status)
}
