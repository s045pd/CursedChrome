// Package busx provides a thin pub/sub abstraction for cross-process bot
// routing. It mirrors the Redis channels used by the Node.js server:
//
//	TOBROWSER_<browser_id>  server -> bot RPC requests
//	TOPROXY_<browser_id>    bot    -> server RPC replies
//	SYSTEM_EVENTS           system-wide broadcasts
//
// A "fake" in-memory backend is provided for unit tests and smoke runs.
package busx

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Channel name prefixes. Kept identical to server.js to maintain
// wire-compatibility with any existing infrastructure.
const (
	ChanToBrowserPrefix = "TOBROWSER_"
	ChanToProxyPrefix   = "TOPROXY_"
	ChanSystemEvents    = "SYSTEM_EVENTS"
)

// ToBrowserChan returns the redis channel name for outbound RPC to a bot.
func ToBrowserChan(browserID string) string { return ChanToBrowserPrefix + browserID }

// ToProxyChan returns the redis channel name for replies from a bot.
func ToProxyChan(browserID string) string { return ChanToProxyPrefix + browserID }

// Bus is the minimal interface all backends satisfy.
type Bus interface {
	Publish(ctx context.Context, channel string, payload []byte) error
	Subscribe(ctx context.Context, channels ...string) (Subscription, error)
	Close() error
}

// Message carries a single inbound payload.
type Message struct {
	Channel string
	Payload []byte
}

// Subscription is a stream of inbound messages on the channels
// requested. Cancel by calling Close.
type Subscription interface {
	Channel() <-chan Message
	Close() error
}

// ErrClosed is returned by Subscribe / Publish after Close.
var ErrClosed = errors.New("bus closed")

// MemoryBus is an in-process implementation suitable for unit tests
// and the SKIP_REDIS smoke mode.
type MemoryBus struct {
	mu      sync.RWMutex
	subs    map[string]map[*memSub]struct{}
	closed  bool
}

// NewMemoryBus returns a ready MemoryBus.
func NewMemoryBus() *MemoryBus {
	return &MemoryBus{subs: make(map[string]map[*memSub]struct{})}
}

// Publish delivers payload to every subscriber of channel synchronously
// (non-blocking on slow consumers — slow ones drop the message).
func (m *MemoryBus) Publish(_ context.Context, channel string, payload []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return ErrClosed
	}
	for sub := range m.subs[channel] {
		select {
		case sub.ch <- Message{Channel: channel, Payload: payload}:
		default:
			// drop on full buffer; matches "best-effort pubsub"
		}
	}
	return nil
}

// Subscribe creates a subscription joining the given channels.
func (m *MemoryBus) Subscribe(_ context.Context, channels ...string) (Subscription, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil, ErrClosed
	}
	s := &memSub{
		bus:      m,
		ch:       make(chan Message, 64),
		channels: append([]string(nil), channels...),
	}
	for _, ch := range channels {
		set, ok := m.subs[ch]
		if !ok {
			set = make(map[*memSub]struct{})
			m.subs[ch] = set
		}
		set[s] = struct{}{}
	}
	return s, nil
}

// Close releases all subscribers.
func (m *MemoryBus) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	for _, set := range m.subs {
		for s := range set {
			close(s.ch)
		}
	}
	m.subs = nil
	return nil
}

type memSub struct {
	bus      *MemoryBus
	ch       chan Message
	channels []string
	once     sync.Once
}

func (s *memSub) Channel() <-chan Message { return s.ch }
func (s *memSub) Close() error {
	s.once.Do(func() {
		s.bus.mu.Lock()
		defer s.bus.mu.Unlock()
		if s.bus.subs == nil {
			return
		}
		for _, ch := range s.channels {
			delete(s.bus.subs[ch], s)
		}
		// Drain to prevent goroutine leaks
		go func() { //nolint:revive // intentional drain goroutine
			for range s.ch {
			}
		}()
		close(s.ch)
	})
	return nil
}

// Format helpers for debugging.
func (m Message) String() string { return fmt.Sprintf("Message(%s, %d bytes)", m.Channel, len(m.Payload)) }
