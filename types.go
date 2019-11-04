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
	AddFunc(f ...FuncStack)
	Count() uint64
	Drain(onChannel bool) ([]interface{}, error)
	FuncMergeBytes() FuncStack
	FuncMergeChunkedElements() FuncStack
	FuncReturnIndividualBytes() FuncStack
	GetChan() (<-chan interface{}, error)
	GetOnChan(amount int) error
	Get(amount int) ([]interface{}, error)
	Unshift(objs ...interface{}) error
	Ping()
	Purge() error
	Push(objs ...interface{}) error
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
	// default gate amount, open up the gate is this amount of elements need to be drained. (useful in conjunction with callback functions)
	GateAmount int64
	// debug output of the drain handlers' current elements
	Debug bool
}
