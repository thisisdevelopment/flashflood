package flashflood_test

import (
	"testing"

	"github.com/thisisdevelopment/flashflood/v2"
)

// TestObj and getTestObjs are defined in flashflood__01_test.go

func TestPushAndDeferClose(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
	})
	defer ff.Close()

	o := getTestObjs(5)

	ch, err := ff.GetChan()

	_ = ff.Push(o[0], o[1], o[2], o[3], o[4])

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

	_ = ff.Push(o[4])
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

func TestPushAndClose(t *testing.T) {
	ff := flashflood.New[TestObj](&flashflood.Opts{
		BufferAmount: 3,
	})

	o := getTestObjs(5)

	ch, err := ff.GetChan()

	_ = ff.Push(o[0], o[1], o[2], o[3], o[4])

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

	_ = ff.Push(o[4])
	select {
	case v := <-ch:
		// fmt.Println("V", v)
		if v.Key != "k2" || v.Value != "v2" {
			t.Fatalf("expected: %#v; got %#v", o[1], v)
		}
	default:
		t.Fatalf("expected: %#v; got nothing", o[0])
	}

	_, _ = ff.Drain(false, false)
	ff.Close()
}
