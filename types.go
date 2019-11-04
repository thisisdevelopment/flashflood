package flashflood

import (
	"sync"
	"time"
)

// FuncStack type function to be called as callback on drained elements
type FuncStack func(objs []interface{}, ff *FlashFlood) []interface{}

//FlashFlood struct
type FlashFlood struct {
	buffer          []interface{}
	bufferAmount    int64
	mutex           *sync.Mutex
	ticker          *time.Ticker
	floodChan       chan interface{}
	channelFetched  bool
	lastAction      time.Time
	lastActionMutex *sync.Mutex
	timeout         time.Duration

	funcstack  []FuncStack
	gateAmount int64
	debug      bool
}

//FF the interface
type FF interface {
	Push(objs ...interface{}) error
	Unshift(objs ...interface{}) error
	Drain(onChannel bool) ([]interface{}, error)
	GetChan() (<-chan interface{}, error)
	Ping()
	Count() uint64
	Purge() error
	AddFunc(f ...FuncStack)
	FuncMergeBytes() FuncStack
	FuncReturnIndividualBytes() FuncStack
	FuncMergeChunkedElements() FuncStack
}

//Opts ...
type Opts struct {
	// the amount of the internal buffer, if buffer is full elements will be drained to channel
	BufferAmount int64
	// default time before the buffer times out and will start draining its contents to the channel
	Timeout time.Duration
	// default ticker time the buffer will check for activity (see Timeout)
	TickerTime time.Duration
	// the amount the channel will buffer
	ChannelBuffer uint64
	// default gate amount, open up the gate is this amount of elements need to be drained. (useful in conjuction with callback functions)
	GateAmount int64
	// debug output of the drain handlers' current elements
	Debug bool
}
