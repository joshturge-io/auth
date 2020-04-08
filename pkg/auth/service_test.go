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
	srv        *auth.Service
	password   = "123password"
	cipherKeys = []string{
		"vcMGBMVbxobHRRdX1WBYq0T4L3UYWQLd",
		"EvMT3FFDNX9dW3SggfyC7sJJ74EkzH32",
		"tHWYreQPuHhfPLIIqcAliQWgfXdNVWLF",
	}
)

func resetRepo() {
	repository.TestUser = map[string]string{
		"salt":       "25b072f201ef24e750dcc558eaf2d8f3",
		"hash":       "1743545c93d519060a72e5671a66cbe898163b41d8be2a92a57ac3b6a2650c8394cf4f009aa0df642721145694879ace89c1a9973ff601538220d6a59f665524022fc789a3f6512d7f4654ff8f39c7ba7ec5b12e93c08df97be9f8a4",
		"refresh":    "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
		"expiration": strconv.FormatInt(time.Now().Add(24*time.Hour).Unix(), 10),
	}

	repository.TestBlacklist = []string{}
}

func init() {
	repo := repository.NewTestRepository()
	srv = auth.NewService("secret", repo, cipherKeys, &auth.Options{
		RefreshTokenLength:     32,
		JWTokenExpiration:      15 * time.Minute,
		RefreshTokenExpiration: 24 * time.Hour,
		SaltLength:             16,
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
