package token_test

import (
	"testing"
	"time"

	"github.com/joshturge-io/auth/pkg/token"
)

var secret = "secret"

func TestGenerate(t *testing.T) {
	jw := token.NewJW(secret, "", 15*time.Minute)
	if err := jw.Generate(); err != nil {
		t.Error(err)
	}
}
