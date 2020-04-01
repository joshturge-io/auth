package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var ErrJWInvalid = errors.New("token is invalid")

type authClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

type JW struct {
	secret, username string
	exp              time.Time
	tokenStr         string
}

func NewJW(secret, username string, exp time.Duration) *JW {
	return &JW{secret, username, time.Now().Add(exp), ""}
}

func NewJWFromExisting(secret, tokenStr string) (*JW, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v: %w", token.Header["alg"], ErrJWInvalid)
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %s: %w", tokenStr, err)
	}

	if !token.Valid {
		return nil, ErrJWInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims: token claims type: %T: %w", claims, ErrJWInvalid)
	}

	t := &JW{secret: secret, tokenStr: tokenStr}

	t.username, ok = claims["username"].(string)
	if !ok {
		return nil, fmt.Errorf("failed to extract username from claim: wanted: string got: %T: %w",
			claims["username"], ErrJWInvalid)
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("failed to extract expiration from claim: wanted: float64 got: %T: %w",
			claims["exp"], ErrJWInvalid)
	}

	t.exp = time.Unix(int64(exp), 0)

	return t, nil
}

func (t *JW) Token() string {
	return t.tokenStr
}

func (t *JW) Generate() error {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &authClaims{
		Username: t.username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: t.exp.Unix(),
		},
	})

	tokenStr, err := token.SignedString([]byte(t.secret))
	if err != nil {
		return fmt.Errorf("could not sign jwt: %w", err)
	}

	t.tokenStr = tokenStr

	return nil
}

func (t *JW) ExpiresIn() time.Duration {
	return time.Until(t.exp)
}

func (t *JW) IsExpired() bool {
	return t.exp.Unix() < time.Now().Unix()
}
