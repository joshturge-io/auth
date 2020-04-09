package redis_test

import (
	"os"
	"testing"

	"github.com/go-redis/redis"
	"github.com/joshturge-io/auth/pkg/flush"
	redisFlush "github.com/joshturge-io/auth/pkg/flush/redis"
)

var flusher flush.Flusher

func init() {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REPO_ADDR"),
		Password: os.Getenv("REPO_PSWD"),
		DB:       0,
	})
	flusher = redisFlush.NewRedisFlusher(client)
}

func TestFlush(t *testing.T) {
	if err := flusher.Flush(); err != nil {
		t.Error(err)
		t.FailNow()
	}
}
