package repository

import (
	"context"
	"sort"
	"strconv"
	"time"
)

var TestUser = map[string]string{
	"salt":       "25b072f201ef24e750dcc558eaf2d8f3",
	"hash":       "1743545c93d519060a72e5671a66cbe898163b41d8be2a92a57ac3b6a2650c8394cf4f009aa0df642721145694879ace89c1a9973ff601538220d6a59f665524022fc789a3f6512d7f4654ff8f39c7ba7ec5b12e93c08df97be9f8a4",
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
