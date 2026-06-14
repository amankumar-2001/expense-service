// Package cacherepo holds Redis-backed caches. The committed-money summary is
// read on every dashboard load, so it is cached with a short TTL and invalidated
// whenever the user changes an autopay or their salary. This uses the same Redis
// instance as auth-service (a separate DB index isolates the keyspace).
package cacherepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ISummaryCache caches the committed-money summary per user.
type ISummaryCache interface {
	// GetCommitted returns the cached JSON payload and true on a hit.
	GetCommitted(ctx context.Context, userID int64) ([]byte, bool, error)
	// SetCommitted stores the JSON payload with the given TTL.
	SetCommitted(ctx context.Context, userID int64, data []byte, ttl time.Duration) error
	// Invalidate drops the cached summary (called after any mutation).
	Invalidate(ctx context.Context, userID int64) error
}

type summaryCache struct {
	rdb *redis.Client
}

// NewSummaryCache constructs the Redis-backed summary cache.
func NewSummaryCache(rdb *redis.Client) ISummaryCache {
	return &summaryCache{rdb: rdb}
}

func committedKey(userID int64) string {
	return fmt.Sprintf("expense:committed:user:%d", userID)
}

func (c *summaryCache) GetCommitted(ctx context.Context, userID int64) ([]byte, bool, error) {
	v, err := c.rdb.Get(ctx, committedKey(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get committed cache: %w", err)
	}
	return v, true, nil
}

func (c *summaryCache) SetCommitted(ctx context.Context, userID int64, data []byte, ttl time.Duration) error {
	if ttl <= 0 {
		return nil // caching disabled
	}
	if err := c.rdb.Set(ctx, committedKey(userID), data, ttl).Err(); err != nil {
		return fmt.Errorf("set committed cache: %w", err)
	}
	return nil
}

func (c *summaryCache) Invalidate(ctx context.Context, userID int64) error {
	if err := c.rdb.Del(ctx, committedKey(userID)).Err(); err != nil {
		return fmt.Errorf("invalidate committed cache: %w", err)
	}
	return nil
}
