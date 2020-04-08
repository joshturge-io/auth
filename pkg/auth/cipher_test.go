package auth_test

import (
	"encoding/hex"
	"testing"

	"github.com/joshturge-io/auth/pkg/auth"
)

var (
	chall *auth.Challenger
	keys  = [][]byte{
		[]byte("vcMGBMVbxobHRRdX1WBYq0T4L3UYWQLd"),
		[]byte("EvMT3FFDNX9dW3SggfyC7sJJ74EkzH32"),
		[]byte("tHWYreQPuHhfPLIIqcAliQWgfXdNVWLF"),
	}
	user = map[string]string{
		"password": "123password",
	}
)

func init() {
	chall = auth.NewChallenger(16, keys)
}

func TestGenerate(t *testing.T) {
	var err error
	user["salt"], user["cipher"], err = chall.Generate(user["password"])
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	t.Logf("%s %s\n", user["salt"], user["cipher"])

	if ciph, err := hex.DecodeString(user["cipher"]); len(ciph) != 92 {
		if err != nil {
			t.Error(err)
		}
		t.Errorf("len of cipher is not 92 got: %d\n", len(ciph))
	}

	if salt, err := hex.DecodeString(user["salt"]); len(salt) != 16 {
		if err != nil {
			t.Error(err)
		}
		t.Errorf("len of salt is not 16 got: %d\n", len(salt))
	}
}

func TestValidate(t *testing.T) {
	valid, err := chall.Validate(user["salt"], user["password"], user["cipher"])
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	if !valid {
		t.Error("cipher is not valid")
	}
}
