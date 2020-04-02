package repository

import (
	"context"
	"io"
	"time"
)

type Withdrawer interface {
	GetRefreshToken(userId string) (string, error)
	GetSalt(userId string) (string, error)
	GetHash(userId string) (string, error)
}

type Depositor interface {
	SetRefreshToken(userId string, token string, exp time.Duration) error
	SetSalt(userId string, salt string) error
	SetHash(userId string, hash string) error
}

type DepositWithdrawer interface {
	Withdrawer
	Depositor
	SetBlacklist(token string, exp time.Duration) error
	IsBlacklisted(token string) (bool, error)
	RemoveRefreshToken(userId string) error
	WithContext(ctx context.Context)
}

type Repository interface {
	io.Closer
	DepositWithdrawer
}
