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
	run := true

	wg := &sync.WaitGroup{}
	wg.Add(amount)
	errChan := make(chan error)

	go func(run *bool, errChan *chan error) {
		for *run {
			select {

			case v := <-ch:
				// spew.Dump(v)
				// panic(nil)
				cnt += len(v.([]interface{}))
				// fmt.Println("LEN IS ", len(v.([]interface{})), cnt)
				wg.Add(len(v.([]interface{})) * -1)

				if cnt == amount {
					*run = false
					close(*errChan)
					break
				}
			case <-ctx.Done():
				*errChan <- fmt.Errorf("expected: counter of %d ; got counter of %d", amount, cnt)
			}
		}
	}(&run, &errChan)

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
	run := true

	wg := &sync.WaitGroup{}
	wg.Add(amount)

	errChan := make(chan error)

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	go func(run *bool, errChan *chan error) {
		for *run {
			select {

			case v := <-ch:
				// spew.Dump(v)
				// panic(nil)
				cnt += len(v.([]interface{}))
				// fmt.Println("LEN IS ", len(v.([]interface{})), cnt)
				wg.Add(len(v.([]interface{})) * -1)
				if cnt == amount {
					*run = false
					close(*errChan)
					break
				}
			case <-ctx.Done():
				*errChan <- fmt.Errorf("expected: counter of %d ; got counter of %d", amount, cnt)
			}
		}
	}(&run, &errChan)

	select {
	case chanErr := <-errChan:
		if chanErr != nil {
			t.Fatalf(chanErr.Error())
		}
	}

	wg.Wait()
}
