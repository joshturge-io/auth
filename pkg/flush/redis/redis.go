package redis

import (
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/joshturge-io/auth/pkg/flush"
)

// redisFlush contains methods that satisfy the Flusher interface
type redisFlush struct {
	*redis.Client
}

// NewRedisFlusher creates a new Flusher for a redis server
func NewRedisFlusher(client *redis.Client) flush.Flusher {
	return &redisFlush{client}
}

// Flush a redis database blacklist of all expired tokens
func (rf *redisFlush) Flush() error {
	if err := rf.ZRemRangeByScore("blacklist", "0",
		strconv.FormatInt(time.Now().Unix(), 10)).Err(); err != nil {
		return fmt.Errorf("unable to flush blacklist: %w", err)
	}
	return nil
}
