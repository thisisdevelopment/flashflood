package flashflood_test

import (
	"testing"

	"github.com/thisisdevelopment/flashflood"
)

func TestPushAndDeferClose(t *testing.T) {
	ff := flashflood.New(&flashflood.Opts{
		BufferAmount: 3,
	})
	defer ff.Close()

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

func TestPushAndClose(t *testing.T) {
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

	ff.Drain(false, false)
	ff.Close()

}
