package flashflood

import (
	"context"
	"sync"
	"time"
)

// FuncStack type function to be called as callback on drained elements
type FuncStack[T any] func(objs []T, ff *FlashFlood[T]) []T

// FlashFlood struct with generic type parameter
type FlashFlood[T any] struct {
	buffer       []T
	bufferAmount int64
	mutex        *sync.Mutex

	tickerCtx    context.Context
	tickerCancel *context.CancelFunc
	ticker       *time.Ticker
	tickerWg     *sync.WaitGroup

	floodChan      chan T
	channelFetched *ChannelFetchedStatus

	lastAction *sync.Map
	lastFlush  *sync.Map

	flushTimeout time.Duration
	flushEnabled bool
	timeout      time.Duration

	funcstack  []FuncStack[T]
	gateAmount int64
	debug      bool
	opts       *Opts
}

// FF the generic interface
type FF[T any] interface {
	AddFunc(f FuncStack[T])
	Close()
	Count() uint64
	Drain(onChannel bool, respectGate bool) ([]T, error)
	GetChan() (<-chan T, error)
	GetOnChan(amount int) error
	Get(amount int) ([]T, error)
	Unshift(objs ...T) error
	Ping()
	Purge() error
	Push(objs ...T) error
}

// Opts ...
type Opts struct {
	// the amount of the internal buffer, if buffer is full elements will be drained to channel
	BufferAmount int64
	// disable Ring functionality until channel is fetched. Beaware Buffer will increase until we can flush it to the channel
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
