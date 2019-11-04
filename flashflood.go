// Package flashflood is a ringbuffer on steroids.
//
//The buffer is a traditional ring buffer but has features to receive the flushed out elements on a channel.
//You can set the gate amount in order to group elements flushed out and/or perform transformations on them using FuncStack callbacks
package flashflood

import (
	"fmt"
	"sync"
	"time"
)

const (
	// the amount the channel will buffer
	defaultChannelBuffer = 4096
	// the amount of the internal buffer, if buffer is full elements will be drained to channel
	defaultBufferAmount = 256
	// default time before the buffer times out and will start draining its contents to the channel
	defaultTimeout = time.Duration(100 * time.Millisecond)
	// default ticker time the buffer will check for activity (see defaultTimeout )
	defaultTickerTime = time.Duration(10 * time.Millisecond)
	// default gate amount, open up the gate is this amount of elements need to be drained. (useful in conjunction with callback functions)
	defaultGateAmount = int64(1)
	// debug output of the drain handlers' current elements
	debug = false
)

//New returns new instance
func New(opts *Opts) *FlashFlood {

	opts = handleOpts(opts)

	ff := &FlashFlood{
		bufferAmount:    opts.BufferAmount,
		channelFetched:  false,
		debug:           opts.Debug,
		floodChan:       make(chan interface{}, opts.ChannelBuffer),
		funcstack:       []FuncStack{debugFunc},
		gateAmount:      opts.GateAmount,
		lastAction:      time.Now(),
		lastActionMutex: &sync.Mutex{},

		flushTimeout:   opts.FlushTimeout,
		lastFlushMutex: &sync.Mutex{},
		flushEnabled:   opts.FlushEnabled,

		mutex:   &sync.Mutex{},
		ticker:  time.NewTicker(opts.TickerTime),
		timeout: opts.Timeout,
	}

	if ff.flushEnabled {
		ff.lastFlush = time.Now()
	}

	go handleTicker(ff)
	return ff
}

func handleOpts(opts *Opts) *Opts {
	defaultOpts := &Opts{
		BufferAmount:  defaultBufferAmount,
		ChannelBuffer: defaultChannelBuffer,
		Debug:         false,
		GateAmount:    defaultGateAmount,
		Timeout:       defaultTimeout,
		TickerTime:    defaultTickerTime,
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

func handleTicker(ff *FlashFlood) {
	for {
		select {
		case <-ff.ticker.C:
			ff.lastActionMutex.Lock()
			elapsed := time.Since(ff.lastAction)
			ff.lastActionMutex.Unlock()

			if elapsed > ff.timeout {
				ff.Drain(true)
			} else {
				if ff.flushEnabled {
					ff.lastFlushMutex.Lock()
					elapsed = time.Since(ff.lastFlush)
					ff.lastFlushMutex.Unlock()

					if elapsed > ff.flushTimeout {
						ff.Drain(true)
					}
				}
			}
		}
	}
}

func (i *FlashFlood) handleDrainObjs() []interface{} {

	var drainObjs []interface{}
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

//Push add objects to buffer
func (i *FlashFlood) Push(objs ...interface{}) error {
	i.mutex.Lock()
	i.buffer = append(i.buffer, objs...)
	drainObjs := i.handleDrainObjs()
	i.flush2Channel(drainObjs, false)
	i.mutex.Unlock()
	i.Ping()
	// TODO err is always nil here, stub to not break bw compatibility in the future
	return nil
}

//Unshift add objects to the front of buffer
func (i *FlashFlood) Unshift(objs ...interface{}) error {
	i.mutex.Lock()
	i.buffer = append(objs, i.buffer...)
	drainObjs := i.handleDrainObjs()
	i.flush2Channel(drainObjs, false)
	i.mutex.Unlock()
	i.Ping()
	// TODO err is always nil here, stub to not break bw compatibility in the future
	return nil
}

func (i *FlashFlood) flush2Channel(objs []interface{}, isInteralBuffer bool) {

	bl := int64(len(objs))

	if bl > 0 && i.channelFetched {

		if isInteralBuffer {

			if i.gateAmount > 1 {
				if bl > i.gateAmount {
					objs, i.buffer = i.buffer[0:i.gateAmount], i.buffer[i.gateAmount:]
				} else {
					objs = i.buffer
					i.clearBuffer()
				}
			} else {
				objs = i.buffer
				i.clearBuffer()
			}

		}

		for _, f := range i.funcstack {
			objs = f(objs, i)
		}

		for _, v := range objs {
			i.floodChan <- v
		}

		if isInteralBuffer && len(i.buffer) > 0 {
			i.flush2Channel(i.buffer, true)
		}

	}

}

//GetChan get the overflow channel
func (i *FlashFlood) GetChan() (<-chan interface{}, error) {
	i.channelFetched = true
	// TODO err is always nil here, stub to not break bw compatibility in the future
	return i.floodChan, nil
}

// Purge clears buffer
func (i *FlashFlood) Purge() error {
	i.mutex.Lock()
	i.clearBuffer()
	i.mutex.Unlock()
	return nil
}

//Ping updates lastaction to postpone timeout
func (i *FlashFlood) Ping() {
	i.lastActionMutex.Lock()
	i.lastAction = time.Now()
	i.lastActionMutex.Unlock()
}

//Count returns amount of elements in buffer
func (i *FlashFlood) Count() uint64 {
	i.mutex.Lock()
	cnt := uint64(len(i.buffer))
	i.mutex.Unlock()
	return cnt
}

func (i *FlashFlood) clearBuffer() {
	// make sure we have a mutex Lock
	i.Ping()
	i.buffer = nil
}

//Get amount of elements from buffer
func (i *FlashFlood) Get(amount int) ([]interface{}, error) {
	var drainObjs []interface{}

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

//GetOnChan amount of elements from buffer, flush to channel
func (i *FlashFlood) GetOnChan(amount int) error {
	drainObjs, err := i.Get(amount)
	_ = err
	//TODO implement error handling once Get can throw an error

	i.flush2Channel(drainObjs, false)

	return nil
}

//Drain drains buffer into channel or as slice (onChannel bool)
func (i *FlashFlood) Drain(onChannel bool) ([]interface{}, error) {

	i.lastFlushMutex.Lock()
	i.lastFlush = time.Now()
	i.lastFlushMutex.Unlock()

	i.mutex.Lock()
	if len(i.buffer) == 0 {
		i.mutex.Unlock()
		return nil, nil
	}

	if onChannel {
		i.flush2Channel(i.buffer, true)
		i.clearBuffer()
		i.mutex.Unlock()
		return nil, nil
	}

	objs := i.buffer
	i.clearBuffer()
	i.mutex.Unlock()

	for _, f := range i.funcstack {
		objs = f(objs, i)
	}

	return objs, nil
}

//AddFunc add a "callback" function to the callstack to be performed on the objects drained
func (i *FlashFlood) AddFunc(f FuncStack) {
	l := len(i.funcstack)
	debughandler := i.funcstack[l-1]
	i.funcstack[l-1] = f
	i.funcstack = append(i.funcstack, debughandler)
}

func debugFunc(i []interface{}, ff *FlashFlood) []interface{} {
	// fmt.Println("IN DEFAULT FUNC")
	if ff.debug {
		fmt.Printf("DEBUG: %#v\n", i)
	}
	return i
}
