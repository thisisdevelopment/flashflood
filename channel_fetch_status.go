package flashflood

import "sync"

const fk = "fetched"

// ChannelFetchedStatus helper to determine if flush channel was fetched by instance
type ChannelFetchedStatus interface {
	ChannelFetched()
	IsChannelFetched() bool
}

// CFS implementation struct helper to determine if flush channel was fetched by instance
type CFS struct {
	fetched *sync.Map
}

// NewChannelFetchedStatus returns new instance
func NewChannelFetchedStatus() ChannelFetchedStatus {
	o := &CFS{
		fetched: &sync.Map{},
	}
	return o
}

// ChannelFetched set the flush channel as fetched
func (s *CFS) ChannelFetched() {
	s.fetched.Store(fk, true)
}

// IsChannelFetched flush channel is fetched
func (s *CFS) IsChannelFetched() bool {
	if _, ok := s.fetched.Load(fk); ok {
		return true
	}
	return false
}
