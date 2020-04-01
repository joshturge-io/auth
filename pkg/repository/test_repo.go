package repository

import (
	"context"
	"sort"
	"strconv"
	"time"
)

var TestUser = map[string]string{
	"salt":       "H4jk53hGsk3fj4Dfsj3",
	"hash":       "dd373f6f7e9338d82a5ccab1be65475c06e97fed63cd59b892024a0a120aa6f0",
	"refresh":    "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
	"expiration": strconv.FormatInt(time.Now().Add(24*time.Hour).Unix(), 10),
}

var TestBlacklist = []string{}

type testRepository struct{}

func NewTestRepository() Repository {
	return &testRepository{}
}

func (tr *testRepository) GetRefreshToken(TestUserId string) (string, error) {
	return TestUser["refresh"], nil
}

func (tr *testRepository) GetSalt(TestUserId string) (string, error) {
	return TestUser["salt"], nil
}

func (tr *testRepository) GetHash(TestUserId string) (string, error) {
	return TestUser["hash"], nil
}

func (tr *testRepository) SetRefreshToken(TestUserId, token string, exp time.Duration) error {
	TestUser["refresh"] = token
	TestUser["expiration"] = strconv.FormatInt(time.Now().Add(exp).Unix(), 10)
	return nil
}

func (tr *testRepository) SetSalt(TestUserId, salt string) error {
	TestUser["salt"] = salt
	return nil
}

func (tr *testRepository) SetHash(TestUserId, hash string) error {
	TestUser["hash"] = hash
	return nil
}

func (tr *testRepository) SetBlacklist(token string, exp time.Duration) error {
	TestBlacklist = append(TestBlacklist, token)
	return nil
}

func (tr *testRepository) IsBlacklisted(token string) (bool, error) {
	sort.Strings(TestBlacklist)
	index := sort.SearchStrings(TestBlacklist, token)

	return index < len(TestBlacklist) && TestBlacklist[index] == token, nil
}

func (tr *testRepository) RemoveRefreshToken(TestUserId string) error {
	TestUser["refresh"] = ""
	TestUser["expiration"] = ""
	return nil
}

func (tr *testRepository) WithContext(ctx context.Context) {}

func (tr *testRepository) Close() error {
	TestUser = nil
	TestBlacklist = []string{}
	return nil
}
