package flashflood

import (
	"testing"
)

func TestDefaultOpts(t *testing.T) {
	ff := New[string](&Opts{
		BufferAmount: 3,
	})

	if ff.bufferAmount != 3 {
		t.Fatalf("expected: %d; got %d", 3, ff.bufferAmount)
	}

	if ff.timeout != defaultTimeout {
		t.Fatalf("expected: %v; got %v", defaultTimeout, ff.timeout)
	}
}

func TestChannelFetchedStatus(t *testing.T) {
	o := NewChannelFetchedStatus()
	if o.IsChannelFetched() == true {
		t.Fatalf("expected: false got true")
	}
	o.ChannelFetched()
	if o.IsChannelFetched() == false {
		t.Fatalf("expected: true got false")
	}
}
