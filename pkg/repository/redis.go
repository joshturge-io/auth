package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-redis/redis"
)

// redisKeyStore satisfies the Repository interface
type redisKeyStore struct {
	*redis.Client
}

// NewRedisRepository will create a new connection to a redis server
func NewRedisRepository(addr, password string) (Repository, error) {
	rks := &redisKeyStore{redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})}

	if err := rks.Ping().Err(); err != nil {
		return nil, err
	}

	return rks, nil
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

	if rks.HExists(userId, "refresh").Val() {
		exp, err := rks.HGet(userId, "expiration").Int64()
		if err != nil {
			return token, err
		}

		if exp < time.Now().Unix() {
			if err = rks.RemoveRefreshToken(userId); err != nil {
				return token, err
			}
			return token, ErrNotExist
		}

		token, err = rks.HGet(userId, "refresh").Result()
		if err != nil {
			if errors.Is(err, redis.Nil) {
				return token, ErrNotExist
			}
			return token, err
		}
	}

	return token, ErrNotExist
}

func (rks *redisKeyStore) GetSalt(userId string) (string, error) {
	userId = rks.fmtUserId(userId)

	salt, err := rks.HGet(userId, "salt").Result()
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

	hash, err := rks.HGet(userId, "hash").Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNotExist
		}
		return "", err
	}

	return hash, nil
}

func (rks *redisKeyStore) SetRefreshToken(userId, token string, exp time.Duration) (err error) {
	userId = rks.fmtUserId(userId)

	_, err = rks.Pipelined(func(pipe redis.Pipeliner) error {
		if err = pipe.HSet(userId, "expiration", exp).Err(); err != nil {
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
	return rks.HSet(userId, "salt", salt).Err()
}

func (rks *redisKeyStore) SetHash(userId, hash string) error {
	userId = rks.fmtUserId(userId)
	return rks.HSet(userId, "hash", hash).Err()
}

func (rks *redisKeyStore) IsBlacklisted(token string) (bool, error) {
	_, err := rks.ZRank("blacklist", token).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (rks *redisKeyStore) SetBlacklist(token string, exp time.Duration) error {
	return rks.ZAdd("blacklist", redis.Z{
		Score:  float64(time.Now().Add(exp).Unix()),
		Member: token,
	}).Err()
}

func (rks *redisKeyStore) RemoveRefreshToken(userId string) error {
	userId = rks.fmtUserId(userId)
	return rks.HDel(userId, "refresh", "expiration").Err()
}

func (rks *redisKeyStore) WithContext(ctx context.Context) {
	rks.WithContext(ctx)
}
