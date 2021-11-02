package flashflood

import (
	"testing"
)

func TestDefaultOpts(t *testing.T) {
	ff := New(&Opts{
		BufferAmount: 3,
	})

	if ff.GetBufferAmount() != 3 {
		t.Fatalf("expected: %d; got %d", 3, ff.GetBufferAmount())
	}

	if ff.GetTimeout() != defaultTimeout {
		t.Fatalf("expected: %v; got %v", defaultTimeout, ff.GetTimeout())
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
