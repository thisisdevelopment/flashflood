package flashflood_test

import (
	"fmt"
	"math"
	"os"
	"testing"
	"time"

	"github.com/thisisdevelopment/flashflood/v2"
)

type TestObj struct {
	Key   string
	Value string
}

func getTestObjs(amount int) []TestObj {
	r := []TestObj{}
	for i := 1; i <= amount; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		t := TestObj{key, value}
		r = append(r, t)
	}
	return r
}

func TestPush(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
	})
	o := getTestObjs(5)

	ch, err := ff.GetChan()

	ff.Push(o[0], o[1], o[2], o[3], o[4])

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	select {
	case v := <-ch:
		// fmt.Println("V", v)
		if v.Key != "k1" || v.Value != "v1" {
			t.Fatalf("expected: %#v; got %#v", o[0], v)
		}
	default:
		t.Fatalf("expected: %#v; got nothing", o[0])
	}

	ff.Push(o[4])
	select {
	case v := <-ch:
		// fmt.Println("V", v)
		if v.Key != "k2" || v.Value != "v2" {
			t.Fatalf("expected: %#v; got %#v", o[1], v)
		}
	default:
		t.Fatalf("expected: %#v; got nothing", o[0])
	}
}

func BenchmarkPushChan(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  3,
		ChannelBuffer: uint64(b.N),
	})

	_, _ = ff.GetChan()
	o := getTestObjs(1)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		ff.Push(o[0])
	}
}

func BenchmarkPushNoChan(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:      10 * time.Millisecond,
		BufferAmount: 3,
	})

	// ff.GetChan()
	o := getTestObjs(1)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		ff.Push(o[0])
	}
}

func BenchmarkPushChanGate(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  3,
		ChannelBuffer: uint64(b.N),
		GateAmount:    2,
	})

	_, _ = ff.GetChan()
	o := getTestObjs(1)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		ff.Push(o[0])
	}
}

func BenchmarkPushChanBigBuffer(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  512,
		ChannelBuffer: uint64(b.N),
	})

	_, _ = ff.GetChan()
	o := getTestObjs(1)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		ff.Push(o[0])
	}
}

func BenchmarkWithGet(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  512,
		ChannelBuffer: uint64(b.N),
	})

	_, _ = ff.GetChan()
	o := getTestObjs(1)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// fmt.Println(i)
		ff.Push(o[0])
		_, _ = ff.Get(1)
	}
}

func BenchmarkPushChanBigBufferPow(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  512,
		ChannelBuffer: uint64(b.N),
	})

	ch, _ := ff.GetChan()
	o := getTestObjs(1)
	b.StartTimer()
	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))

		b.Run(fmt.Sprintf("pow/%d", n), func(b *testing.B) {
			go func() {
				for range ch {
				}
			}()
			for i := 0; i < b.N; i++ {
				// fmt.Println(i)
				ff.Push(o[0])
			}
		})
	}
}

func BenchmarkPushChanBigBufferPowCBFuncWithGate(b *testing.B) {
	b.StopTimer()

	ff := flashflood.New[TestObj](&flashflood.Opts{
		Timeout:       10 * time.Millisecond,
		BufferAmount:  512,
		ChannelBuffer: uint64(b.N),
		GateAmount:    3,
	})

	ch, _ := ff.GetChan()
	// Removed FuncMergeChunkedElements - deprecated in v2
	o := getTestObjs(1)
	b.StartTimer()
	for k := 0.; k <= 10; k++ {
		n := int(math.Pow(2, k))

		b.Run(fmt.Sprintf("pow/%d", n), func(b *testing.B) {
			go func() {
				for range ch {
				}
			}()
			for i := 0; i < b.N; i++ {
				// fmt.Println(i)
				ff.Push(o[0])
			}
		})
	}
}

func TestPushDefaultOpts(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{})
	ch, err := ff.GetChan()

	for _, t := range getTestObjs(512) {
		ff.Push(t)
	}

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}
	for i := 1; i <= 512; i++ {
		v := <-ch
		// fmt.Println("V", v)
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		if v.Key != key || v.Value != value {
			t.Fatalf("expected: %s %s; got %#v", key, value, v)
		}
	}
}

func TestPushDefaultOptsWithGate(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		GateAmount: 8,
	})
	ch, err := ff.GetChan()

	for i := 0; i < 512; i++ {
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		testObj := TestObj{key, value}
		ff.Push(testObj)
	}

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}
	for i := 0; i < 512; i++ {
		v := <-ch
		key := fmt.Sprintf("k%d", i)
		value := fmt.Sprintf("v%d", i)
		if v.Key != key || v.Value != value {
			t.Fatalf("expected: %s %s; got %#v", key, value, v)
		}
	}
}

func TestDrainOnChannel(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})

	o := getTestObjs(4)

	ch, err := ff.GetChan()
	ff.Push(o[0], o[1], o[2], o[3])

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	_, _ = ff.Drain(true, false)

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}

	drained := []TestObj{}
	run := true
	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	if drained[0].Key != "k1" || drained[0].Value != "v1" {
		t.Fatalf("expected: %#v; got %#v", o[0], drained[0])
	}
	if drained[1].Key != "k2" || drained[1].Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[1])
	}
	if drained[2].Key != "k3" || drained[2].Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[2])
	}
	if drained[3].Key != "k4" || drained[3].Value != "v4" {
		t.Fatalf("expected: %#v; got %#v", o[3], drained[3])
	}
}

func TestDrainWithReturn(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})
	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	drained, err := ff.Drain(false, false)
	if err != nil {
		t.Fatalf("could not drain object: %v", err)
	}

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}

	if drained[0].Key != "k2" || drained[0].Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[0])
	}
	if drained[1].Key != "k3" || drained[1].Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[1])
	}
	if drained[2].Key != "k4" || drained[2].Value != "v4" {
		t.Fatalf("expected: %#v; got %#v", o[3], drained[2])
	}
}

func TestManyLongTimeout(t *testing.T) {
	maxValue := 100
	bufferAmount := int64(3)
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: bufferAmount,
		Timeout:      1 * time.Second,
		// TickerTime:   time.Duration(2 * time.Second),
	})

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	go func() {
		for _, t := range getTestObjs(maxValue) {
			ff.Push(t)
		}
	}()

	// Reduce sleep time for CI-friendly behavior
	sleepDuration := 100 * time.Millisecond // Default: CI-friendly
	if longTests := os.Getenv("LONG_TESTS"); longTests == "true" {
		sleepDuration = 500 * time.Millisecond // Original timing for thorough testing
	}
	
	time.Sleep(sleepDuration)
	drained := []TestObj{}
	run := true

	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	if int64(len(drained)) != int64(maxValue)-bufferAmount {
		t.Fatalf("expected %d in buffer; got %v", int64(maxValue)-bufferAmount, len(drained))
	}
}

func TestManyShortTimeout(t *testing.T) {
	maxValue := 100

	bufferAmount := int64(3)
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: bufferAmount,
	})

	ch, err := ff.GetChan()
	_ = err

	go func() {
		for _, t := range getTestObjs(maxValue) {
			ff.Push(t)
		}
	}()

	time.Sleep(50 * time.Millisecond)
	drained := []TestObj{}
	run := true
	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	if int64(len(drained)) != int64(maxValue)-bufferAmount {
		t.Fatalf("expected %d in buffer; got %v", int64(maxValue)-bufferAmount, len(drained))
	}
}

func TestPurge(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})

	o := getTestObjs(4)

	_, err := ff.GetChan()
	ff.Push(o[0], o[1], o[2], o[3])

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	_ = ff.Purge()
	// ff.Drain(true)

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}
}

func TestGet(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})
	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	drained, err := ff.Get(2)
	if err != nil {
		t.Fatalf("could not drain object: %v", err)
	}

	if drained[0].Key != "k2" || drained[0].Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[0])
	}
	if drained[1].Key != "k3" || drained[1].Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[1])
	}

	if ff.Count() != 1 {
		t.Fatalf("expected 1 in buffer; got %v", ff.Count())
	}
}

func TestGetWithFunc(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})
	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	f1 := func(objs []TestObj, _ *flashflood.FlashFlood[TestObj]) []TestObj {
		var c []TestObj
		for _, v := range objs {
			// Transform the string representation into a new TestObj
			transformed := TestObj{
				Key:   fmt.Sprintf("---%s", v.Key),
				Value: fmt.Sprintf("---%s", v.Value),
			}
			c = append(c, transformed)
		}
		return c
	}

	ff.AddFunc(f1)

	drained, err := ff.Get(2)
	if err != nil {
		t.Fatalf("could not drain object: %v", err)
	}

	if drained[0].Key != "---k2" || drained[0].Value != "---v2" {
		t.Fatalf("expected: TestObj{Key: '---k2', Value: '---v2'}; got %#v", drained[0])
	}
	if drained[1].Key != "---k3" || drained[1].Value != "---v3" {
		t.Fatalf("expected: TestObj{Key: '---k3', Value: '---v3'}; got %#v", drained[1])
	}

	if ff.Count() != 1 {
		t.Fatalf("expected 1 in buffer; got %v", ff.Count())
	}
}

func TestGetOnChan(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      250 * time.Millisecond,
	})

	ch, _ := ff.GetChan()

	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	_ = ff.GetOnChan(2)

	drained := []TestObj{}
	run := true

	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	if drained[0].Key != "k1" || drained[0].Value != "v1" {
		t.Fatalf("expected: %#v; got %#v", o[0], drained[0])
	}
	if drained[1].Key != "k2" || drained[1].Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[1])
	}
	if drained[2].Key != "k3" || drained[2].Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[2])
	}

	if ff.Count() != 1 {
		t.Fatalf("expected 1 in buffer; got %v", ff.Count())
	}
	// Clear out buffer, and to hit code to test
	rest, _ := ff.Get(10)

	if rest[0].Key != "k4" || rest[0].Value != "v4" {
		t.Fatalf("expected: %#v; got %#v", o[3], rest[0])
	}

	rest, _ = ff.Get(10)
	if rest != nil {
		t.Fatalf("expected rest to be nil; got %#v", rest)
	}
}

func TestFlushTimeout(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 4,
		GateAmount:   2,
		Timeout:      1 * time.Second,
		TickerTime:   5 * time.Millisecond,
		FlushEnabled: true,
		FlushTimeout: 50 * time.Millisecond,
	})

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	o := getTestObjs(5)

	ff.Push(o[0])
	time.Sleep(25 * time.Millisecond)
	ff.Push(o[1])
	time.Sleep(100 * time.Millisecond)

	drained := []TestObj{}
	run := true
	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	if int64(len(drained)) != 2 {
		t.Fatalf("(1) expected %d in buffer; got %v", 2, len(drained))
	}

	ff.Push(o[2])
	ff.Push(o[3])
	ff.Push(o[4])

	// Reduce sleep time for CI-friendly behavior
	sleepDuration := 200 * time.Millisecond // Default: CI-friendly
	if longTests := os.Getenv("LONG_TESTS"); longTests == "true" {
		sleepDuration = 1500 * time.Millisecond // Original timing for thorough testing
	}
	
	time.Sleep(sleepDuration)
	drained = []TestObj{}
	run = true
	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v)
		default:
			run = false
		}
	}

	// fmt.Println(drained)
	if int64(len(drained)) != 3 {
		t.Fatalf("(2) expected %d in buffer; got %v", 3, len(drained))
	}
}
