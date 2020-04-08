package auth

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"

	"golang.org/x/crypto/pbkdf2"
)

var (
	ErrCipherTooShort = errors.New("cipher is too short")
	ErrMessAuthFailed = errors.New("cipher: message authentication failed")
)

// generateRandBytes given the length of the byte slice
func generateRandBytes(n int) ([]byte, error) {
	b := make([]byte, n)

	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to read random crypto bytes: %w", err)
	}

	return b, nil
}

// generateChallengeHash from a salt and password
func generateChallengeHash(salt, password []byte) []byte {
	return pbkdf2.Key(password, salt, 4096, 64, sha256.New)
}

// Challenger holds methods to create and validate user challenges
type Challenger struct {
	saltLen    int
	cipherKeys [][]byte
}

// NewChallenger will initialise a new Challenger
func NewChallenger(saltLen int, keys [][]byte) *Challenger {
	return &Challenger{saltLen, keys}
}

// chooseRandomKeyIndex from the keys provided
func (c *Challenger) chooseRandomKeyIndex() (int, error) {
	keyI, err := rand.Int(rand.Reader, big.NewInt(int64(len(c.cipherKeys))))
	if err != nil {
		return 0, fmt.Errorf("failed to generate a random key index: %w", err)
	}

	return int(keyI.Int64()), nil
}

// encrypt data provided with a cipher key
func (c *Challenger) encrypt(data []byte, keyIndex int) ([]byte, error) {
	block, err := aes.NewCipher(c.cipherKeys[keyIndex])
	if err != nil {
		return nil, fmt.Errorf("failed to create new cipher for key: %x: %w",
			c.cipherKeys[keyIndex], err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm AEAD: %w", err)
	}

	nounce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nounce); err != nil {
		return nil, fmt.Errorf("unable to read in random bytes: %w", err)
	}

	return gcm.Seal(nounce, nounce, data, nil), nil
}

func (c *Challenger) decrypt(data []byte, keyIndex int) ([]byte, error) {
	block, err := aes.NewCipher(c.cipherKeys[keyIndex])
	if err != nil {
		return nil, fmt.Errorf("failed to decode cipher with key: %x: %w",
			c.cipherKeys[keyIndex], err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create new gcm AEAD: %w", err)
	}

	nounceSize := gcm.NonceSize()
	if len(data) < nounceSize {
		return nil, ErrCipherTooShort
	}

	nounce, data := data[:nounceSize], data[nounceSize:]

	return gcm.Open(nil, nounce, data, nil)
}

// Validate a challenge by checking if we can recreate the cipher from the salt and password
// provided
func (c *Challenger) Validate(salt, password, cipherStr string) (bool, error) {
	saltBytes, err := hex.DecodeString(salt)
	if err != nil {
		return false, fmt.Errorf("failed to decode salt: %w", err)
	}

	cipher, err := hex.DecodeString(cipherStr)
	if err != nil {
		return false, fmt.Errorf("failed to decode cipher string: %w", err)
	}

	hashSum := generateChallengeHash(saltBytes, []byte(password))

	for i, _ := range c.cipherKeys {
		decCiph, err := c.decrypt(cipher, i)
		if err != nil {
			if errors.Is(err, ErrMessAuthFailed) {
				continue
			}
			return false, fmt.Errorf("failed to decrypt hash sum: %w", err)
		}

		if bytes.Compare(decCiph, hashSum) == 0 {
			return true, nil
		}
	}

	return false, nil
}

// Generate a new password cipher using a random salt and key
func (c *Challenger) Generate(password string) (salt string, cipher string, err error) {
	randBytes, err := generateRandBytes(c.saltLen)
	if err != nil {
		return "", "", err
	}

	keyIndex, err := c.chooseRandomKeyIndex()
	if err != nil {
		return "", "", nil
	}

	ciph, err := c.encrypt(generateChallengeHash(randBytes, []byte(password)), keyIndex)
	if err != nil {
		return "", "", fmt.Errorf("failed to encrypt hash sum: %w", err)
	}

	return hex.EncodeToString(randBytes), hex.EncodeToString(ciph), nil
}
