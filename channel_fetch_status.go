package flashflood

import "sync"

type ChannelFetchedStatus interface {
	ChannelFetched()
	IsChannelFetched() bool
}

type CFS struct {
	mutex   *sync.Mutex
	fetched bool
}

func NewChannelFetchedStatus() ChannelFetchedStatus {
	o := &CFS{
		mutex:   &sync.Mutex{},
		fetched: false,
	}
	return o
}

func (s *CFS) ChannelFetched() {
	s.mutex.Lock()
	s.fetched = true
	s.mutex.Unlock()
}

func (s *CFS) IsChannelFetched() bool {
	s.mutex.Lock()
	r := s.fetched
	s.mutex.Unlock()
	return r
}
