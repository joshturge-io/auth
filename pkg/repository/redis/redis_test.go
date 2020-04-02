package redis_test

import (
	"os"
	"testing"
	"time"

	"github.com/joshturge-io/auth/pkg/repository"
	"github.com/joshturge-io/auth/pkg/repository/redis"
	"github.com/joshturge-io/auth/pkg/token"
)

var (
	repo          repository.Repository
	testUser      map[string]string
	testBlacklist []string
)

func init() {
	var err error
	repo, err = redis.NewRepository(os.Getenv("REDIS_ADDR"), os.Getenv("REDIS_PSWD"))
	if err != nil {
		panic(err)
	}

	testUser = map[string]string{
		"salt":    "H4jk53hGsk3fj4Dfsj3",
		"hash":    "dd373f6f7e9338d82a5ccab1be65475c06e97fed63cd59b892024a0a120aa6f0",
		"refresh": "Uq_XJB5p5clZ_lAjFVND0oTYT9uFe8plBfGHFGMZ4RI=",
	}

	testBlacklist = []string{}
}

func TestSetRefreshToken(t *testing.T) {
	if err := repo.SetRefreshToken("test_user", testUser["refresh"], 3*time.Minute); err != nil {
		t.Error(err)
	}
}

func TestGetRefreshToken(t *testing.T) {
	token, err := repo.GetRefreshToken("test_user")
	if err != nil {
		t.Error(err)
	}

	if token != testUser["refresh"] {
		t.Errorf("token does not match the one set wanted: %s got: %s\n", testUser["refresh"], token)
	}
}

func TestSetSalt(t *testing.T) {
	if err := repo.SetSalt("test_user", testUser["salt"]); err != nil {
		t.Error(err)
	}
}

func TestGetSalt(t *testing.T) {
	salt, err := repo.GetSalt("test_user")
	if err != nil {
		t.Error(err)
	}

	if salt != testUser["salt"] {
		t.Errorf("salt does not match the one set wanted: %s got: %s", testUser["salt"], salt)
	}
}

func TestSetHash(t *testing.T) {
	if err := repo.SetSalt("test_user", testUser["hash"]); err != nil {
		t.Error(err)
	}
}

func TestGetHash(t *testing.T) {
	hash, err := repo.GetSalt("test_user")
	if err != nil {
		t.Error(err)
	}

	if hash != testUser["hash"] {
		t.Errorf("hash does not match the one set wanted: %s got: %s", testUser["hash"], hash)
	}
}

func TestSetBlacklist(t *testing.T) {
	jw := token.NewJW("secret", "test_user", 3*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}

	testBlacklist = append(testBlacklist, jw.Token())

	if err := repo.SetBlacklist(jw.Token(), 3*time.Minute); err != nil {
		t.Error(err)
	}
}

func TestIsBlacklisted(t *testing.T) {
	blacklisted, err := repo.IsBlacklisted(testBlacklist[0])
	if err != nil {
		t.Error(err)
	}

	if !blacklisted {
		t.Error("token was not blacklisted")
	}
}
