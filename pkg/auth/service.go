package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joshturge-io/auth/pkg/repository"
	"github.com/joshturge-io/auth/pkg/repository/redis"
	"github.com/joshturge-io/auth/pkg/token"
	"golang.org/x/sync/errgroup"
)

var (
	ErrUserNotExist     = errors.New("user does not exist")
	ErrInvalidChallenge = errors.New("user provided incorrect challenge")
	ErrInvalidSession   = errors.New("session is not valid")
)

// Session holds information about users session
type Session struct {
	UserId  string
	Refresh string
	JWT     string
}

// Options for tokens
type Options struct {
	// Token Length for a refresh token
	RefreshTokenLength int
	// JWT Expiration time
	JWTokenExpiration time.Duration
	// Refresh Expiration time
	RefreshTokenExpiration time.Duration
	// length of password salts
	SaltLength int
}

// Service is an authentication service used for manipulating sessions
type Service struct {
	repo      repository.DepositWithdrawer
	chall     *Challenger
	jwtSecret string
	opt       *Options
}

// NewService will create a new auth service
func NewService(secret string, repo repository.DepositWithdrawer, keys []string,
	opt *Options) *Service {
	cipherKeys := make([][]byte, len(keys))
	for i, key := range keys {
		cipherKeys[i] = []byte(key)
	}

	return &Service{repo, NewChallenger(opt.SaltLength, cipherKeys), secret, opt}
}

// generateSession will generate a new session
func (s *Service) generateSession(ctx context.Context, userId string) (*Session, error) {
	s.repo.WithContext(ctx)
	var (
		ref = make(chan string, 1)
		jwt = make(chan string, 1)
	)
	defer func() {
		close(ref)
		close(jwt)
	}()

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		refresh, err := token.GenerateRefresh(s.opt.RefreshTokenLength)
		if err != nil {
			return fmt.Errorf("could not generate refresh token: %w", err)
		}

		ref <- refresh

		return nil
	})
	errs.Go(func() error {
		jw := token.NewJW(s.jwtSecret, userId, s.opt.JWTokenExpiration)
		if err := jw.Generate(); err != nil {
			return fmt.Errorf("failed to generate jwt: %w", err)
		}

		jwt <- jw.Token()

		return nil
	})

	if err := errs.Wait(); err != nil {
		if errors.Is(err, redis.ErrNotExist) {
			return nil, ErrUserNotExist
		}
		return nil, err
	}

	refresh := <-ref

	if err := s.repo.SetRefreshToken(userId, refresh,
		s.opt.RefreshTokenExpiration); err != nil {
		return nil, fmt.Errorf("could not set refresh token: %w", err)
	}

	return &Session{
		Refresh: refresh,
		JWT:     <-jwt,
	}, nil
}

// validateSession will make sure a session has a valid refresh and jw token
func (s *Service) validateSession(ctx context.Context, sess *Session) error {
	switch {
	case sess.UserId == "":
		return ErrInvalidSession
	case sess.Refresh == "":
		return ErrInvalidSession
	case sess.JWT == "":
		return ErrInvalidSession
	}

	validity := make(chan bool, 2)
	defer close(validity)

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		valid, err := s.IsValidRefresh(ctx, sess.UserId, sess.Refresh)
		if err != nil {
			return err
		}

		validity <- valid

		return nil
	})
	errs.Go(func() error {
		blacklisted, err := s.repo.IsBlacklisted(sess.JWT)
		if err != nil {
			return fmt.Errorf("unable to check blacklist status of token: %w", err)
		}
		valid, err := s.IsValidJWT(sess.JWT)
		if err != nil {
			return err
		}

		validity <- valid && !blacklisted

		return nil
	})

	if err := errs.Wait(); err != nil {
		return fmt.Errorf("failed to validate: %w", err)
	}

	// check if either goroutine found and returned an invalid token
	if !<-validity || !<-validity {
		return ErrInvalidSession
	}

	return nil
}

// SessionWithChallenge create a new session provided a valid challenge (username and password)
func (s *Service) SessionWithChallenge(ctx context.Context, userId, password string) (*Session,
	error) {
	if userId == "" || password == "" {
		return nil, ErrInvalidChallenge
	}

	s.repo.WithContext(ctx)
	var (
		saltChan    = make(chan string, 1)
		hashChan    = make(chan string, 1)
		sessionChan = make(chan *Session, 1)
	)
	defer func() {
		close(saltChan)
		close(hashChan)
		close(sessionChan)
	}()

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		salt, err := s.repo.GetSalt(userId)
		if err != nil {
			return fmt.Errorf("could not get salt for user: %s from repository: %w", userId,
				err)
		}

		saltChan <- salt

		return nil
	})
	errs.Go(func() error {
		hash, err := s.repo.GetHash(userId)
		if err != nil {
			return fmt.Errorf("could not get hash for user: %s from repository: %w", userId,
				err)
		}

		hashChan <- hash

		return nil
	})
	errs.Go(func() error {
		session, err := s.generateSession(ctx, userId)
		if err != nil {
			return err
		}

		sessionChan <- session

		return nil
	})

	if err := errs.Wait(); err != nil {
		if errors.Is(err, redis.ErrNotExist) {
			return nil, ErrUserNotExist
		}
		return nil, err
	}

	if valid, err := s.chall.Validate(<-saltChan, password, <-hashChan); !valid {
		if err != nil {
			return nil, fmt.Errorf("failed to validate challenge: %w", err)
		}

		return nil, ErrInvalidChallenge
	}

	return <-sessionChan, nil
}

// IsValidRefresh will query the repository and validate that it exists, if it doesn't then the
// token is invalid
func (s *Service) IsValidRefresh(ctx context.Context, userId, refresh string) (bool, error) {
	s.repo.WithContext(ctx)
	userRefresh, err := s.repo.GetRefreshToken(userId)
	if err != nil {
		if errors.Is(err, redis.ErrNotExist) {
			return false, ErrUserNotExist
		}
		return false, fmt.Errorf("unable to get refresh token for userId: %s: %w", userId, err)
	}

	return userRefresh == refresh, nil
}

// IsValidJWT will attempt to parse the jwt and check if it has expired. If it fails to parse
// or has expired then the jwt is invalid
func (s *Service) IsValidJWT(tokenStr string) (bool, error) {
	t, err := token.NewJWFromExisting(s.jwtSecret, tokenStr)
	if err != nil {
		return false, err
	}

	return !t.IsExpired(), nil
}

// DestroySession will invalidate a session by blacklisting the jwt and removing the refresh
// token from the users record
func (s *Service) DestroySession(ctx context.Context, old *Session) error {
	s.repo.WithContext(ctx)
	if err := s.validateSession(ctx, old); err != nil {
		return err
	}

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		if err := s.repo.RemoveRefreshToken(old.UserId); err != nil {
			if errors.Is(err, redis.ErrNotExist) {
				return ErrUserNotExist
			}
			return fmt.Errorf("failed to remove refresh token: %w", err)
		}

		return nil
	})
	errs.Go(func() error {
		jw, err := token.NewJWFromExisting(s.jwtSecret, old.JWT)
		if err != nil {
			return err
		}
		if err := s.repo.SetBlacklist(jw.Token(), jw.ExpiresIn()); err != nil {
			return fmt.Errorf("could not blacklist token %w", err)
		}

		return nil
	})

	return errs.Wait()
}

// Renew will delete a current users session and generate them a new one
func (s *Service) Renew(ctx context.Context, old *Session) (*Session, error) {
	if old.UserId == "" || old.Refresh == "" {
		return nil, ErrInvalidSession
	}

	errs, ctx := errgroup.WithContext(ctx)
	if old.JWT != "" {
		errs.Go(func() error {
			jw, err := token.NewJWFromExisting(s.jwtSecret, old.JWT)
			if err != nil {
				return err
			}

			if err := s.repo.SetBlacklist(jw.Token(), jw.ExpiresIn()); err != nil {
				return fmt.Errorf("could not blacklist token %w", err)
			}

			return nil
		})
	}
	errs.Go(func() error {
		if valid, err := s.IsValidRefresh(ctx, old.UserId, old.Refresh); !valid {
			if err != nil {
				return err
			}
		}

		if err := s.repo.RemoveRefreshToken(old.UserId); err != nil {
			if errors.Is(err, redis.ErrNotExist) {
				return ErrUserNotExist
			}
			return fmt.Errorf("could not remove refresh token: %w", err)
		}

		return nil
	})

	if err := errs.Wait(); err != nil {
		return nil, err
	}

	return s.generateSession(ctx, old.UserId)
}
