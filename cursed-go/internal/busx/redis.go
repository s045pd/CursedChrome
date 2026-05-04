package busx

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisBus is a redis-backed Bus implementation.
type RedisBus struct {
	client *redis.Client
}

// NewRedisBus connects to redis at host:port and returns a Bus.
func NewRedisBus(ctx context.Context, host string, port int) (*RedisBus, error) {
	cli := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", host, port),
	})
	if err := cli.Ping(ctx).Err(); err != nil {
		_ = cli.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return &RedisBus{client: cli}, nil
}

// Publish forwards to redis PUBLISH.
func (r *RedisBus) Publish(ctx context.Context, channel string, payload []byte) error {
	return r.client.Publish(ctx, channel, payload).Err()
}

// Subscribe wraps a redis PubSub as Subscription.
func (r *RedisBus) Subscribe(ctx context.Context, channels ...string) (Subscription, error) {
	ps := r.client.Subscribe(ctx, channels...)
	if _, err := ps.Receive(ctx); err != nil {
		_ = ps.Close()
		return nil, err
	}
	out := make(chan Message, 64)
	go func() {
		defer close(out)
		for msg := range ps.Channel() {
			select {
			case out <- Message{Channel: msg.Channel, Payload: []byte(msg.Payload)}:
			case <-ctx.Done():
				return
			}
		}
	}()
	return &redisSub{ps: ps, out: out}, nil
}

// Close closes the redis client.
func (r *RedisBus) Close() error { return r.client.Close() }

type redisSub struct {
	ps  *redis.PubSub
	out chan Message
}

func (s *redisSub) Channel() <-chan Message { return s.out }
func (s *redisSub) Close() error             { return s.ps.Close() }
