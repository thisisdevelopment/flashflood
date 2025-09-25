// Package flashflood is a ringbuffer on steroids.
//
// The buffer is a traditional ring buffer but has features to receive the flushed out elements on a channel.
// You can set the gate amount in order to group elements flushed out and/or perform transformations on them using FuncStack callbacks
package flashflood

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

const (
	// the amount the channel will buffer
	defaultChannelBuffer = 4096
	// the amount of the internal buffer, if buffer is full elements will be drained to channel
	defaultBufferAmount = 256
	// default time before the buffer times out and will start draining its contents to the channel
	defaultTimeout = 100 * time.Millisecond
	// default ticker time the buffer will check for activity (see defaultTimeout )
	defaultTickerTime = 10 * time.Millisecond
	// default gate amount, open up the gate is this amount of elements need to be drained. (useful in conjunction with callback functions)
	defaultGateAmount = int64(1)
)

const (
	lastAction = iota
	lastFlush
)

// New returns new instance with generic type parameter
func New[T any](opts *Opts) *FlashFlood[T] {
	opts = handleOpts(opts)
	nfs := NewChannelFetchedStatus()

	tickerCtx, tickerCancel := context.WithCancel(context.Background())
	var tickerWg sync.WaitGroup

	ff := &FlashFlood[T]{
		bufferAmount:   opts.BufferAmount,
		channelFetched: &nfs,
		debug:          opts.Debug,
		floodChan:      make(chan T, opts.ChannelBuffer),
		funcstack:      []FuncStack[T]{debugFunc[T]},
		gateAmount:     opts.GateAmount,

		lastAction: &sync.Map{},

		flushTimeout: opts.FlushTimeout,

		flushEnabled: opts.FlushEnabled,

		mutex:        &sync.Mutex{},
		tickerCtx:    tickerCtx,
		tickerCancel: &tickerCancel,
		ticker:       time.NewTicker(opts.TickerTime),
		tickerWg:     &tickerWg,
		timeout:      opts.Timeout,
		opts:         opts,

		lastFlush: &sync.Map{},
	}
	ff.lastAction.Store(lastAction, time.Now())

	if ff.flushEnabled {
		ff.lastFlush.Store(lastFlush, time.Now())
	}

	// Start ticker goroutine after all initialization is complete
	go handleTicker[T](ff)
	return ff
}

func handleOpts(opts *Opts) (defaultOpts *Opts) {
	defaultOpts = &Opts{
		BufferAmount:               defaultBufferAmount,
		ChannelBuffer:              defaultChannelBuffer,
		Debug:                      false,
		GateAmount:                 defaultGateAmount,
		Timeout:                    defaultTimeout,
		TickerTime:                 defaultTickerTime,
		DisableRingUntilChanActive: false,
	}

	if opts.ChannelBuffer == 0 {
		opts.ChannelBuffer = defaultOpts.ChannelBuffer
	}

	if opts.BufferAmount == 0 {
		opts.BufferAmount = defaultOpts.BufferAmount
	}
	if opts.Timeout == 0 {
		opts.Timeout = defaultOpts.Timeout
	}

	if opts.TickerTime == 0 {
		opts.TickerTime = defaultOpts.TickerTime
	}

	if opts.GateAmount == 0 {
		opts.GateAmount = defaultOpts.GateAmount
	}

	return opts
}

// Close Cleanup resources and kill timers/tickers etc
func (i *FlashFlood[T]) Close() {
	// Stop ticker and wait for goroutine to finish
	(*i.tickerCancel)()
	i.tickerWg.Wait()

	i.channelFetched = nil
	i.floodChan = nil

	i.funcstack = nil
	i.lastAction = nil

	i.opts = nil

	i.mutex = nil

	if len(i.buffer) != 0 {
		log.Println("Close called on non empty buffer")
	}

	i.buffer = nil
}

func handleTicker[T any](i *FlashFlood[T]) {
	i.tickerWg.Add(1)
	defer i.tickerWg.Done()

	run := true
	var elapsed time.Duration

	for run {
		select {
		case <-i.tickerCtx.Done():
			run = false
		case <-i.ticker.C:

			if e, ok := i.lastAction.Load(lastAction); ok {
				elapsed = time.Since(e.(time.Time))
			}

			if elapsed > i.timeout {
				_, _ = i.Drain(true, false)
			} else {
				if i.flushEnabled {

					if e, ok := i.lastFlush.Load(lastFlush); ok {
						elapsed = time.Since(e.(time.Time))
					}

					if elapsed > i.flushTimeout {
						_, _ = i.Drain(true, true)
					}
				}
			}
		}
	}

	i.ticker.Stop()
	i.ticker.C = nil
	i.ticker = nil
}

func (i *FlashFlood[T]) handleDrainObjs() []T {
	if i.opts.DisableRingUntilChanActive && !(*i.channelFetched).IsChannelFetched() {
		return nil
	}

	var drainObjs []T
	bl := int64(len(i.buffer))
	toDrain := bl - i.bufferAmount

	if i.gateAmount == 1 {
		if toDrain > 0 {
			drainObjs, i.buffer = i.buffer[0:toDrain], i.buffer[toDrain:]
		}
	} else {
		if toDrain > 0 && toDrain >= i.gateAmount {
			drainObjs, i.buffer = i.buffer[0:i.gateAmount], i.buffer[i.gateAmount:]
		}
	}

	return drainObjs
}

// Push add objects to buffer
func (i *FlashFlood[T]) Push(objs ...T) error {
	i.mutex.Lock()
	i.buffer = append(i.buffer, objs...)
	drainObjs := i.handleDrainObjs()
	if drainObjs != nil {
		i.flush2Channel(drainObjs, false, false)
	}
	i.mutex.Unlock()
	i.Ping()
	return nil
}

// Unshift add objects to the front of buffer
func (i *FlashFlood[T]) Unshift(objs ...T) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.buffer = append(objs, i.buffer...)
	drainObjs := i.handleDrainObjs()
	if drainObjs != nil {
		i.flush2Channel(drainObjs, false, false)
	}
	i.Ping()
	return nil
}

func (i *FlashFlood[T]) flush2Channel(objs []T, isInteralBuffer bool, respectGate bool) {
	bl := int64(len(objs))

	if bl > 0 && (*i.channelFetched).IsChannelFetched() {

		if isInteralBuffer {
			if i.gateAmount > 1 {
				if bl >= i.gateAmount {
					objs, i.buffer = objs[0:i.gateAmount], objs[i.gateAmount:]
				} else {
					if !respectGate {
						objs = i.buffer
						i.clearBuffer()
					}
				}
			} else {
				objs = i.buffer
				i.clearBuffer()
			}
		}
		blAfter := int64(len(objs))

		for _, f := range i.funcstack {
			objs = f(objs, i)
		}

		for _, v := range objs {
			i.floodChan <- v
		}

		if isInteralBuffer && len(i.buffer) > 0 && blAfter < bl {
			i.flush2Channel(i.buffer, true, respectGate)
		}
	}
}

// GetChan get the overflow channel
func (i *FlashFlood[T]) GetChan() (<-chan T, error) {
	(*i.channelFetched).ChannelFetched()
	return i.floodChan, nil
}

// Purge clears buffer
func (i *FlashFlood[T]) Purge() error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.clearBuffer()
	return nil
}

// Ping updates lastaction to postpone timeout
func (i *FlashFlood[T]) Ping() {
	i.lastAction.Store(lastAction, time.Now())
}

// Count returns amount of elements in buffer
func (i *FlashFlood[T]) Count() uint64 {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	cnt := uint64(len(i.buffer))
	return cnt
}

func (i *FlashFlood[T]) clearBuffer() {
	// make sure we have a mutex Lock
	i.Ping()
	i.buffer = nil
}

// Get amount of elements from buffer
func (i *FlashFlood[T]) Get(amount int) ([]T, error) {
	var drainObjs []T

	i.mutex.Lock()
	bl := len(i.buffer)
	if bl == 0 {
		i.mutex.Unlock()
		return nil, nil
	}

	if bl <= amount {
		objs := i.buffer
		i.clearBuffer()
		i.mutex.Unlock()

		for _, f := range i.funcstack {
			objs = f(objs, i)
		}

		return objs, nil
	}

	drainObjs, i.buffer = i.buffer[0:amount], i.buffer[amount:]
	i.mutex.Unlock()
	for _, f := range i.funcstack {
		drainObjs = f(drainObjs, i)
	}
	return drainObjs, nil
}

// GetOnChan amount of elements from buffer, flush to channel
func (i *FlashFlood[T]) GetOnChan(amount int) error {
	drainObjs, err := i.Get(amount)
	_ = err
	// TODO implement error handling once Get can throw an error

	i.flush2Channel(drainObjs, false, false)

	return nil
}

// Drain drains buffer into channel or as slice (onChannel bool)
func (i *FlashFlood[T]) Drain(onChannel bool, respectGate bool) ([]T, error) {
	i.lastFlush.Store(lastFlush, time.Now())

	i.mutex.Lock()
	defer i.mutex.Unlock()

	if len(i.buffer) == 0 {
		return nil, nil
	}

	if onChannel {
		i.flush2Channel(i.buffer, true, respectGate)
		i.clearBuffer()
		return nil, nil
	}

	objs := i.buffer
	i.clearBuffer()

	for _, f := range i.funcstack {
		objs = f(objs, i)
	}

	return objs, nil
}

// AddFunc add a "callback" function to the callstack to be performed on the objects drained
func (i *FlashFlood[T]) AddFunc(f FuncStack[T]) {
	l := len(i.funcstack)
	debughandler := i.funcstack[l-1]
	i.funcstack[l-1] = f
	i.funcstack = append(i.funcstack, debughandler)
}

func debugFunc[T any](i []T, ff *FlashFlood[T]) []T {
	if ff.debug {
		fmt.Printf("DEBUG: %#v\n", i)
	}
	return i
}
