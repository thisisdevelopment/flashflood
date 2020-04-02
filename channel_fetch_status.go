package flashflood

import "sync"

// ChannelFetchedStatus helper to determine if flush channel was fetched by instance
type ChannelFetchedStatus interface {
	ChannelFetched()
	IsChannelFetched() bool
}

// CFS implementation struct helper to determine if flush channel was fetched by instance
type CFS struct {
	mutex   *sync.Mutex
	fetched bool
}

// NewChannelFetchedStatus returns new instance
func NewChannelFetchedStatus() ChannelFetchedStatus {
	o := &CFS{
		mutex:   &sync.Mutex{},
		fetched: false,
	}
	return o
}

// ChannelFetched set the flush channel as fetched
func (s *CFS) ChannelFetched() {
	s.mutex.Lock()
	s.fetched = true
	s.mutex.Unlock()
}

// IsChannelFetched flush channel is fetched
func (s *CFS) IsChannelFetched() bool {
	s.mutex.Lock()
	r := s.fetched
	s.mutex.Unlock()
	return r
}
