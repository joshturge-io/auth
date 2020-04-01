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
	srv      *auth.Service
	password = "secret6"
)

func resetRepo() {
	repository.TestUser = map[string]string{
		"salt":       "H4jk53hGsk3fj4Dfsj3",
		"hash":       "dd373f6f7e9338d82a5ccab1be65475c06e97fed63cd59b892024a0a120aa6f0",
		"refresh":    "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
		"expiration": strconv.FormatInt(time.Now().Add(24*time.Hour).Unix(), 10),
	}

	repository.TestBlacklist = []string{}
}

func init() {
	repo := repository.NewTestRepository()
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
	session, err := srv.SessionWithChallenge(ctx, "user", password)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("Refresh Token: %s\n", session.Refresh)
	t.Logf("JWT: %s\n", session.JWT)
}

func TestDestroySession(t *testing.T) {
	resetRepo()

	jw := token.NewJW("secret", "user", 15*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	if err := srv.DestroySession(ctx, &auth.Session{"user", repository.TestUser["refresh"],
		jw.Token()}); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if i := sort.SearchStrings(repository.TestBlacklist,
		jw.Token()); repository.TestBlacklist[i] != jw.Token() {
		t.Error("JWT not in repository.TestBlacklist")
	}
}

func TestRenew(t *testing.T) {
	resetRepo()
	jw := token.NewJW("secret", "user", 15*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}
	oldSess := &auth.Session{
		UserId:  "user",
		Refresh: repository.TestUser["refresh"],
		JWT:     jw.Token(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Millisecond)
	defer cancel()

	newSess, err := srv.Renew(ctx, oldSess)
	if err != nil {
		t.Error(err)
	}

	if i := sort.SearchStrings(repository.TestBlacklist,
		jw.Token()); repository.TestBlacklist[i] != jw.Token() {
		t.Error("JWT not in repository.TestBlacklist")
	}

	if repository.TestUser["refresh"] != newSess.Refresh {
		t.Error("new refresh has not been set on repository.TestUser")
	}
}
