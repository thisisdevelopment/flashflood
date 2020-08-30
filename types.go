package flashflood

import (
	"context"
	"sync"
	"time"
)

// FuncStack type function to be called as callback on drained elements
type FuncStack func(objs []interface{}, ff *FlashFlood) []interface{}

//FlashFlood struct
type FlashFlood struct {
	buffer       []interface{}
	bufferAmount int64
	mutex        *sync.Mutex

	tickerCtx    context.Context
	tickerCancel *context.CancelFunc
	ticker       *time.Ticker

	floodChan      chan interface{}
	channelFetched *ChannelFetchedStatus

	lastAction *sync.Map
	lastFlush  *sync.Map

	flushTimeout time.Duration
	flushEnabled bool
	timeout      time.Duration

	funcstack  []FuncStack
	gateAmount int64
	debug      bool
	opts       *Opts
}

//FF the interface
type FF interface {
	AddFunc(f ...FuncStack)
	Close()
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
	//disable Ring functionality until channel is fetched. Beaware Buffer will increase until we can flush it to the channel
	DisableRingUntilChanActive bool
	// enable the flush timeout
	FlushEnabled bool
	// default time before the buffer flush times out and will start draining its contents to the channel
	FlushTimeout time.Duration
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
