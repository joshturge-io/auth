package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/joshturge-io/auth/pkg/auth"
	"github.com/joshturge-io/auth/pkg/grpc"
	"github.com/joshturge-io/auth/pkg/grpc/service"
	"github.com/joshturge-io/auth/pkg/repository"
	"github.com/joshturge-io/auth/pkg/repository/redis"
	"golang.org/x/sync/errgroup"
)

// App holds all the repository and gRPC server methods
type App struct {
	repo repository.Repository
	srv  *grpc.Server
	lg   *log.Logger
}

// Initialise the repository and create the gRPC server
func (a *App) Initialise(configPath string) (err error) {
	a.lg = log.New(os.Stdout, "[INFO] ", log.Ltime|log.Ldate)

	a.lg.Println("Reading configuration file")

	repoPswd := os.Getenv("REPOS_PSWD")
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return errors.New("environment variable JWT_SECRET not set")
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("unable to read configuration: %w", err)
	}

	a.lg.Println("Creating connection to database")

	if os.Getenv("TEST_REPO") != "" {
		a.lg.Println("WARNING: Using test repository")
		a.repo = repository.NewTestRepository()
	} else {
		a.repo, err = redis.NewRepository(a.lg, config.Repo.Address, repoPswd,
			time.Duration(config.Repo.FlushInterval)*time.Second)
		if err != nil {
			return fmt.Errorf("failed to make connection to database: %w", err)
		}
	}

	a.lg.Printf("Creating gRPC server on: %s\n", config.Address)

	opt := &auth.Options{
		RefreshTokenLength:     config.Token.Refresh.Length,
		JWTokenExpiration:      time.Duration(config.Token.Jwt.Expiration) * time.Minute,
		RefreshTokenExpiration: time.Duration(config.Token.Refresh.Expiration) * time.Hour,
		SaltLength:             config.Cipher.SaltLength,
	}

	a.srv, err = grpc.NewServer(config.Address,
		service.NewGRPCAuthService(auth.NewService(jwtSecret, a.repo, config.Cipher.Keys, opt),
			a.lg))
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %w", err)
	}

	return nil
}

// Start serving the gRPC server
func (a *App) Start() error {
	a.srv.Serve()
	a.lg.Println("Started gRPC server")
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt)
	<-sigChan

	return a.srv.Err()
}

// Shutdown the connection to the repository and close the gRPC server
func (a *App) Shutdown(ctx context.Context) error {
	a.lg.Println("Closing connection to database")
	a.lg.Println("Closing gRPC server")

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		return a.srv.Close(ctx)
	})
	errs.Go(a.repo.Close)

	return errs.Wait()
}
