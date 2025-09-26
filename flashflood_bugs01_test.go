package flashflood_test

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/thisisdevelopment/flashflood/v2"
)

// TestObj and getTestObjs are defined in flashflood__01_test.go

// TestLargeWithGate confirmation of bug I found on one of my implementations
func TestLargeWithGateAndFuncScenario1(t *testing.T) {
	ff := flashflood.New[[]TestObj](&flashflood.Opts{
		BufferAmount: 256,
		Timeout:      time.Duration(250 * time.Millisecond),
		FlushTimeout: time.Duration(500 * time.Millisecond),
		FlushEnabled: true,
		GateAmount:   64,
		TickerTime:   time.Duration(50 * time.Millisecond),
	})

	ff.AddFunc(flashflood.FuncMergeChunkedElements[TestObj]())

	// Default to smaller dataset for CI-friendly behavior
	// Only use large dataset when running with -test.long or similar custom flag
	amount := 10000                             // Default: CI-friendly small dataset
	timeoutDuration := 10000 * time.Millisecond // Default: shorter timeout

	// Check for environment variable to run long tests
	if longTests := os.Getenv("LONG_TESTS"); longTests == "true" {
		amount = 100000 // Full scale for thorough testing
		timeoutDuration = 30000 * time.Millisecond
	}

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
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
				cnt += len(v) // v is []TestObj chunk, count all items in it
				// fmt.Println("LEN IS ", len(v), cnt)
				wg.Add(len(v) * -1) // Decrement by number of items in chunk

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
		ff.Push([]TestObj{o}) // Push single-element slice for chunked processing
		// time.Sleep(1 * time.Millisecond)
	}

	select {
	case chanErr := <-errChan:
		if chanErr != nil {
			t.Fatal(chanErr)
		}
	}

	wg.Wait()
}

func TestLargeWithGateAndFuncScenario2(t *testing.T) {
	ff := flashflood.New[[]TestObj](&flashflood.Opts{
		BufferAmount:               256,
		Timeout:                    time.Duration(250 * time.Millisecond),
		FlushTimeout:               500 * time.Millisecond,
		FlushEnabled:               true,
		GateAmount:                 64,
		TickerTime:                 time.Duration(50 * time.Millisecond),
		DisableRingUntilChanActive: true,
	})

	ff.AddFunc(flashflood.FuncMergeChunkedElements[TestObj]())

	// Default to smaller dataset for CI-friendly behavior
	// Only use large dataset when running with LONG_TESTS=true environment variable
	amount := 10000                             // Default: CI-friendly small dataset
	timeoutDuration := 10000 * time.Millisecond // Default: shorter timeout

	// Check for environment variable to run long tests
	if longTests := os.Getenv("LONG_TESTS"); longTests == "true" {
		amount = 100000 // Full scale for thorough testing
		timeoutDuration = 30000 * time.Millisecond
	}

	objs := getTestObjs(amount)
	for _, o := range objs {
		ff.Push([]TestObj{o}) // Push single-element slice for chunked processing
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDuration)
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
				cnt += len(v) // v is []TestObj chunk, count all items in it
				// fmt.Println("LEN IS ", len(v), cnt)
				wg.Add(len(v) * -1) // Decrement by number of items in chunk
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
			t.Fatal(chanErr)
		}
	}

	wg.Wait()
}
