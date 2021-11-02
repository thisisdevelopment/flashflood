package flashflood_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/thisisdevelopment/flashflood"
)

//TestLargeWithGate confirmation of bug I found on one of my implementations
func TestLargeWithGateAndFuncScenario1(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 256,
		Timeout:      time.Duration(250 * time.Millisecond),
		FlushTimeout: time.Duration(500 * time.Millisecond),
		FlushEnabled: true,
		GateAmount:   64,
		TickerTime:   time.Duration(50 * time.Millisecond),
	})

	ff.AddFunc(ff.FuncMergeChunkedElements())

	amount := 100000

	ch, err := ff.GetChan()

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Millisecond)
	defer cancel()

	cnt := 0

	wg := &sync.WaitGroup{}
	wg.Add(amount)
	errChan := make(chan error)

	go func(errChan *chan error) {
	BREAKLOOP:
		for {
			select {

			case v := <-ch:
				cnt += len(v.([]interface{}))
				wg.Add(len(v.([]interface{})) * -1)

				if cnt == amount {
					close(*errChan)
					break BREAKLOOP
				}
			case <-ctx.Done():
				*errChan <- fmt.Errorf("expected: counter of %d ; got counter of %d", amount, cnt)
			}
		}
	}(&errChan)

	objs := getTestObjs(amount)
	for _, o := range objs {
		ff.Push(o)
		// time.Sleep(1 * time.Millisecond)
	}

	select {
	case chanErr := <-errChan:
		if chanErr != nil {
			t.Fatalf(chanErr.Error())
		}
	}

	wg.Wait()

}

func TestLargeWithGateAndFuncScenario2(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount:               256,
		Timeout:                    time.Duration(250 * time.Millisecond),
		FlushTimeout:               time.Duration(500 * time.Millisecond),
		FlushEnabled:               true,
		GateAmount:                 64,
		TickerTime:                 time.Duration(50 * time.Millisecond),
		DisableRingUntilChanActive: true,
	})

	ff.AddFunc(ff.FuncMergeChunkedElements())
	amount := 100000

	objs := getTestObjs(amount)
	for _, o := range objs {
		ff.Push(o)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30000*time.Millisecond)
	defer cancel()

	cnt := 0
	wg := &sync.WaitGroup{}
	wg.Add(amount)

	errChan := make(chan error)

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	go func(errChan *chan error) {
	BREAKLOOP:
		for {
			select {

			case v := <-ch:
				// spew.Dump(v)
				// panic(nil)
				cnt += len(v.([]interface{}))
				// fmt.Println("LEN IS ", len(v.([]interface{})), cnt)
				wg.Add(len(v.([]interface{})) * -1)
				if cnt == amount {
					close(*errChan)
					break BREAKLOOP
				}
			case <-ctx.Done():
				*errChan <- fmt.Errorf("expected: counter of %d ; got counter of %d", amount, cnt)
			}
		}
	}(&errChan)

	select {
	case chanErr := <-errChan:
		if chanErr != nil {
			t.Fatalf(chanErr.Error())
		}
	}
	wg.Wait()
}
