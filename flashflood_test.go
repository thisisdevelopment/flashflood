package flashflood_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/thisisdevelopment/flashflood"
)

type TestObj struct {
	Key   string
	Value string
}

func TestPush(t *testing.T) {
	ff := flashflood.New(&flashflood.Opts{
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
		if v.(TestObj).Key != "k1" || v.(TestObj).Value != "v1" {
			t.Fatalf("expected: %#v; got %#v", o[0], v)
		}
	default:
		t.Fatalf("expected: %#v; got nothing", o[0])
	}

	ff.Push(o[4])
	select {
	case v := <-ch:
		// fmt.Println("V", v)
		if v.(TestObj).Key != "k2" || v.(TestObj).Value != "v2" {
			t.Fatalf("expected: %#v; got %#v", o[1], v)
		}
	default:
		t.Fatalf("expected: %#v; got nothing", o[0])
	}

}

func TestPushDefaultOpts(t *testing.T) {
	ff := flashflood.New(&flashflood.Opts{})
	ch, err := ff.GetChan()

	for _, t := range getTestObjs(512) {
		ff.Push(t)
	}

	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}
	for i := 1; i <= 512; i++ {
		select {
		case v := <-ch:
			// fmt.Println("V", v)
			key := fmt.Sprintf("k%d", i)
			value := fmt.Sprintf("v%d", i)
			if v.(TestObj).Key != key || v.(TestObj).Value != value {
				t.Fatalf("expected: %s %s; got %#v", key, value, v)
			}
		}
	}
}
func TestPushDefaultOptsWithGate(t *testing.T) {
	ff := flashflood.New(&flashflood.Opts{
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
		select {
		case v := <-ch:
			key := fmt.Sprintf("k%d", i)
			value := fmt.Sprintf("v%d", i)
			if v.(TestObj).Key != key || v.(TestObj).Value != value {
				t.Fatalf("expected: %s %s; got %#v", key, value, v)
			}
		}
	}
}

func TestDrainOnChannel(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
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

	ff.Drain(true)

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}

	drained := []TestObj{}
	run := true
	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v.(TestObj))
		default:
			run = false
			break
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

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})
	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	drained, err := ff.Drain(false)
	if err != nil {
		t.Fatalf("could not drain object: %v", err)
	}

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}

	if drained[0].(TestObj).Key != "k2" || drained[0].(TestObj).Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[0])
	}
	if drained[1].(TestObj).Key != "k3" || drained[1].(TestObj).Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[1])
	}
	if drained[2].(TestObj).Key != "k4" || drained[2].(TestObj).Value != "v4" {
		t.Fatalf("expected: %#v; got %#v", o[3], drained[2])
	}

}

func TestManyLongTimeout(t *testing.T) {
	max := 100
	bufferAmount := int64(3)
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: bufferAmount,
		Timeout:      time.Duration(1 * time.Second),
		// TickerTime:   time.Duration(2 * time.Second),
	})

	ch, err := ff.GetChan()
	if err != nil {
		t.Fatalf("could not get channel: %v", err)
	}

	go func() {
		for _, t := range getTestObjs(max) {
			ff.Push(t)
		}
	}()

	time.Sleep(500 * time.Millisecond)
	drained := []TestObj{}
	run := true

	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v.(TestObj))
		default:
			run = false
			break
		}

	}

	if int64(len(drained)) != int64(max)-bufferAmount {
		t.Fatalf("expected %d in buffer; got %v", int64(max)-bufferAmount, len(drained))
	}

}

func TestManyShortTimeout(t *testing.T) {
	max := 100

	bufferAmount := int64(3)
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: bufferAmount,
	})

	ch, err := ff.GetChan()
	_ = err

	go func() {
		for _, t := range getTestObjs(max) {
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
			drained = append(drained, v.(TestObj))
		default:
			run = false
			break
		}

	}

	if int64(len(drained)) != int64(max)-bufferAmount {
		t.Fatalf("expected %d in buffer; got %v", int64(max)-bufferAmount, len(drained))
	}

}

func TestPurge(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
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

	ff.Purge()
	// ff.Drain(true)

	if ff.Count() != 0 {
		t.Fatalf("expected 0 in buffer; got %v", ff.Count())
	}

}

func TestGet(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
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

	if drained[0].(TestObj).Key != "k2" || drained[0].(TestObj).Value != "v2" {
		t.Fatalf("expected: %#v; got %#v", o[1], drained[0])
	}
	if drained[1].(TestObj).Key != "k3" || drained[1].(TestObj).Value != "v3" {
		t.Fatalf("expected: %#v; got %#v", o[2], drained[1])
	}

	if ff.Count() != 1 {
		t.Fatalf("expected 1 in buffer; got %v", ff.Count())
	}

}

func TestGetWithFunc(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})
	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	f1 := func(objs []interface{}, ff *flashflood.FlashFlood) []interface{} {
		var c []interface{}
		for _, v := range objs {
			c = append(c, fmt.Sprintf("---%s_%s", v.(TestObj).Key, v.(TestObj).Value))
		}
		return c
	}

	ff.AddFunc(f1)

	drained, err := ff.Get(2)

	if err != nil {
		t.Fatalf("could not drain object: %v", err)
	}

	if drained[0].(string) != "---k2_v2" {
		t.Fatalf("expected: %#v; got %#v", "---k2_v2", drained[0])
	}
	if drained[1].(string) != "---k3_v3" {
		t.Fatalf("expected: %#v; got %#v", "---k3_v3", drained[1])
	}

	if ff.Count() != 1 {
		t.Fatalf("expected 1 in buffer; got %v", ff.Count())
	}

}

func TestGetOnChan(t *testing.T) {

	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
		Timeout:      time.Duration(250 * time.Millisecond),
	})

	ch, _ := ff.GetChan()

	o := getTestObjs(4)

	ff.Push(o[0], o[1], o[2], o[3])

	if ff.Count() != 3 {
		t.Fatalf("expected 3 in buffer; got %v", ff.Count())
	}

	ff.GetOnChan(2)

	drained := []TestObj{}
	run := true

	for run {
		select {
		case v := <-ch:
			// fmt.Println("V DRAIN", v)
			drained = append(drained, v.(TestObj))
		default:
			run = false
			break
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

	if rest[0].(TestObj).Key != "k4" || rest[0].(TestObj).Value != "v4" {
		t.Fatalf("expected: %#v; got %#v", o[3], rest[0])
	}

	rest, _ = ff.Get(10)
	if rest != nil {
		t.Fatalf("expected rest to be nil; got %#v", rest)
	}

}
