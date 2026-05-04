package busx

import (
	"context"
	"testing"
	"time"
)

func TestMemoryBus_PubSub(t *testing.T) {
	bus := NewMemoryBus()
	defer bus.Close()
	ctx := t.Context()

	sub, err := bus.Subscribe(ctx, "ch1", "ch2")
	if err != nil {
		t.Fatal(err)
	}
	defer sub.Close()

	if err := bus.Publish(ctx, "ch1", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if err := bus.Publish(ctx, "ch2", []byte("world")); err != nil {
		t.Fatal(err)
	}

	got := map[string]string{}
	timeout := time.After(time.Second)
	for i := 0; i < 2; i++ {
		select {
		case m := <-sub.Channel():
			got[m.Channel] = string(m.Payload)
		case <-timeout:
			t.Fatal("timed out waiting for messages")
		}
	}
	if got["ch1"] != "hello" || got["ch2"] != "world" {
		t.Errorf("got = %v", got)
	}
}

func TestMemoryBus_FanOut(t *testing.T) {
	bus := NewMemoryBus()
	defer bus.Close()
	ctx := context.Background()

	a, _ := bus.Subscribe(ctx, "broadcast")
	b, _ := bus.Subscribe(ctx, "broadcast")
	defer a.Close()
	defer b.Close()

	if err := bus.Publish(ctx, "broadcast", []byte("hi")); err != nil {
		t.Fatal(err)
	}

	for i, sub := range []Subscription{a, b} {
		select {
		case m := <-sub.Channel():
			if string(m.Payload) != "hi" {
				t.Errorf("sub %d got %s", i, m.Payload)
			}
		case <-time.After(time.Second):
			t.Errorf("sub %d timed out", i)
		}
	}
}

func TestMemoryBus_UnsubscribeStopsDelivery(t *testing.T) {
	bus := NewMemoryBus()
	defer bus.Close()
	ctx := context.Background()

	sub, _ := bus.Subscribe(ctx, "x")
	sub.Close()

	if err := bus.Publish(ctx, "x", []byte("ignored")); err != nil {
		t.Fatal(err)
	}
	// Just sanity-check we don't panic; delivered to closed channel would panic.
}

func TestMemoryBus_Closed(t *testing.T) {
	bus := NewMemoryBus()
	bus.Close()
	if err := bus.Publish(context.Background(), "x", nil); err != ErrClosed {
		t.Errorf("Publish after close returned %v, want ErrClosed", err)
	}
	if _, err := bus.Subscribe(context.Background(), "x"); err != ErrClosed {
		t.Errorf("Subscribe after close returned %v, want ErrClosed", err)
	}
}

func TestChannelNames(t *testing.T) {
	if ToBrowserChan("abc") != "TOBROWSER_abc" {
		t.Errorf("ToBrowserChan formatted wrong")
	}
	if ToProxyChan("abc") != "TOPROXY_abc" {
		t.Errorf("ToProxyChan formatted wrong")
	}
}
