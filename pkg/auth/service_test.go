package auth_test

import (
	"context"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/joshturge-io/auth/pkg/auth"
	"github.com/joshturge-io/auth/pkg/repository"
	"github.com/joshturge-io/auth/pkg/token"
)

var (
	user      map[string]string
	blacklist []string
	srv       *auth.Service
	password  = "secret6"
)

type testRepository struct{}

func NewTestRepository() repository.DepositWithdrawer {
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

func resetRepo() {
	user = map[string]string{
		"salt":       "H4jk53hGsk3fj4Dfsj3",
		"hash":       "dd373f6f7e9338d82a5ccab1be65475c06e97fed63cd59b892024a0a120aa6f0",
		"refresh":    "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
		"expiration": strconv.FormatInt(time.Now().Add(24*time.Hour).Unix(), 10),
	}

	blacklist = []string{}
}

func init() {
	repo := NewTestRepository()
	srv = auth.NewService("secret", repo, &auth.Options{
		RefreshTokenLength:     32,
		JWTokenExpiration:      15 * time.Minute,
		RefreshTokenExpiration: 24 * time.Hour,
	})

	resetRepo()
}

func TestSessionWithChallenge(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()
	session, err := srv.SessionWithChallenge(ctx, "", password)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("Refresh Token: %s\n", session.Refresh)
	t.Logf("JWT: %s\n", session.JWT)
}

func TestDestroySession(t *testing.T) {
	resetRepo()

	jw := token.NewJW("secret", "", 15*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	if err := srv.DestroySession(ctx, &auth.Session{"", user["refresh"],
		jw.Token()}); err != nil {
		t.Error(err)
	}

	if i := sort.SearchStrings(blacklist, jw.Token()); blacklist[i] != jw.Token() {
		t.Error("JWT not in blacklist")
	}
}

func TestRenew(t *testing.T) {
	resetRepo()
	jw := token.NewJW("secret", "", 15*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}
	oldSess := &auth.Session{
		UserId:  "",
		Refresh: user["refresh"],
		JWT:     jw.Token(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	newSess, err := srv.Renew(ctx, oldSess)
	if err != nil {
		t.Error(err)
	}

	if i := sort.SearchStrings(blacklist, jw.Token()); blacklist[i] != jw.Token() {
		t.Error("JWT not in blacklist")
	}

	if user["refresh"] != newSess.Refresh {
		t.Error("new refresh has not been set on user")
	}
}
