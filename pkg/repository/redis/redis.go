package redis

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/joshturge-io/auth/pkg/flush"
	flusher "github.com/joshturge-io/auth/pkg/flush/redis"
	"github.com/joshturge-io/auth/pkg/repository"
	"golang.org/x/sync/errgroup"
)

var (
	ErrNotExist     = errors.New("member does not exist")
	ErrTokenExpired = errors.New("token has expired")
)

// redisKeyStore satisfies the Repository interface
type redisKeyStore struct {
	client   *redis.Client
	flushSvc *flush.Service
}

// NewRedisRepository will create a new connection to a redis server
func NewRepository(lg *log.Logger, addr, password string,
	flushInt time.Duration) (repository.Repository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		return nil, err
	}

	rks := &redisKeyStore{
		client,
		flush.NewService(lg, flusher.NewRedisFlusher(client), flushInt),
	}

	lg.Println("Starting flushing service")
	rks.flushSvc.Start()

	return rks, rks.flushSvc.Err()
}

// fmtUserId will format a user id to work with are redis repo
func (rks *redisKeyStore) fmtUserId(userId string) string {
	if strings.HasPrefix(userId, "user:") {
		return userId
	}
	return strings.Join([]string{"user", userId}, ":")
}

func (rks *redisKeyStore) GetRefreshToken(userId string) (token string, err error) {
	userId = rks.fmtUserId(userId)

	if rks.client.HExists(userId, "refresh").Val() {
		exp, err := rks.client.HGet(userId, "expiration").Int64()
		if err != nil {
			return token, err
		}

		if exp < time.Now().Unix() {
			if err = rks.RemoveRefreshToken(userId); err != nil {
				return token, err
			}
			return token, ErrTokenExpired
		}

		token, err = rks.client.HGet(userId, "refresh").Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return token, ErrNotExist
			}
			return token, err
		}

		return token, err
	}

	return token, ErrNotExist
}

func (rks *redisKeyStore) GetSalt(userId string) (string, error) {
	userId = rks.fmtUserId(userId)

	salt, err := rks.client.HGet(userId, "salt").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotExist
		}
		return "", err
	}

	return salt, nil
}

func (rks *redisKeyStore) GetHash(userId string) (string, error) {
	userId = rks.fmtUserId(userId)

	hash, err := rks.client.HGet(userId, "hash").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotExist
		}
		return "", err
	}

	return hash, nil
}

func (rks *redisKeyStore) SetRefreshToken(userId, token string,
	exp time.Duration) (err error) {
	userId = rks.fmtUserId(userId)

	_, err = rks.client.Pipelined(func(pipe redis.Pipeliner) error {
		if err = pipe.HSet(userId, "expiration",
			time.Now().Add(exp).Unix()).Err(); err != nil {
			return err
		}
		if err = pipe.HSet(userId, "refresh", token).Err(); err != nil {
			return err
		}

		return nil
	})

	return err
}

func (rks *redisKeyStore) SetSalt(userId, salt string) error {
	userId = rks.fmtUserId(userId)
	return rks.client.HSet(userId, "salt", salt).Err()
}

func (rks *redisKeyStore) SetHash(userId, hash string) error {
	userId = rks.fmtUserId(userId)
	return rks.client.HSet(userId, "hash", hash).Err()
}

func (rks *redisKeyStore) IsBlacklisted(token string) (bool, error) {
	_, err := rks.client.ZRank("blacklist", token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (rks *redisKeyStore) SetBlacklist(token string, exp time.Duration) error {
	return rks.client.ZAdd("blacklist", redis.Z{
		Score:  float64(time.Now().Add(exp).Unix()),
		Member: token,
	}).Err()
}

func (rks *redisKeyStore) RemoveRefreshToken(userId string) error {
	userId = rks.fmtUserId(userId)
	return rks.client.HDel(userId, "refresh", "expiration").Err()
}

func (rks *redisKeyStore) WithContext(ctx context.Context) {
	rks.WithContext(ctx)
}

func (rks *redisKeyStore) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		return rks.flushSvc.Close(ctx)
	})
	errs.Go(rks.client.Close)

	return errs.Wait()
}
