package repository

import (
	"context"
	"sort"
	"strconv"
	"time"
)

var user = map[string]string{
	"salt":       "H4jk53hGsk3fj4Dfsj3",
	"hash":       "dd373f6f7e9338d82a5ccab1be65475c06e97fed63cd59b892024a0a120aa6f0",
	"refresh":    "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
	"expiration": strconv.FormatInt(time.Now().Add(24*time.Hour).Unix(), 10),
}

var blacklist = []string{}

type testRepository struct{}

func NewTestRepository() Repository {
	return &testRepository{}
}

func (tr *testRepository) GetRefreshToken(userId string) (string, error) {
	return user["refresh"], nil
}

func (tr *testRepository) GetSalt(userId string) (string, error) {
	return user["salt"], nil
}

func (tr *testRepository) GetHash(userId string) (string, error) {
	return user["hash"], nil
}

func (tr *testRepository) SetRefreshToken(userId, token string, exp time.Duration) error {
	user["refresh"] = token
	user["expiration"] = strconv.FormatInt(time.Now().Add(exp).Unix(), 10)
	return nil
}

func (tr *testRepository) SetSalt(userId, salt string) error {
	user["salt"] = salt
	return nil
}

func (tr *testRepository) SetHash(userId, hash string) error {
	user["hash"] = hash
	return nil
}

func (tr *testRepository) SetBlacklist(token string, exp time.Duration) error {
	blacklist = append(blacklist, token)
	return nil
}

func (tr *testRepository) IsBlacklisted(token string) (bool, error) {
	sort.Strings(blacklist)
	index := sort.SearchStrings(blacklist, token)

	return index < len(blacklist) && blacklist[index] == token, nil
}

func (tr *testRepository) RemoveRefreshToken(userId string) error {
	user["refresh"] = ""
	user["expiration"] = ""
	return nil
}

func (tr *testRepository) WithContext(ctx context.Context) {}

func (tr *testRepository) Close() error {
	user = nil
	blacklist = []string{}
	return nil
}
