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
	if repoPswd == "" || jwtSecret == "" {
		return errors.New("environment variables REPOS_PSWD or JWT_SECRET not set")
	}

	config, err := ParseConfig(configPath)
	if err != nil {
		return fmt.Errorf("unable to read configuration: %w", err)
	}

	a.lg.Println("Creating connection to database")

	a.repo, err = repository.NewRedisRepository(config.Repo.Address, repoPswd)
	if err != nil {
		return fmt.Errorf("failed to make connection to database: %s", err.Error())
	}

	a.lg.Printf("Creating gRPC server on: %s\n", config.ServerAddr)

	a.srv, err = grpc.NewServer(config.ServerAddr,
		service.NewGRPCAuthService(auth.NewService(jwtSecret, a.repo,
			&auth.Options{
				RefreshTokenLength:     config.Token.RefreshLength,
				JWTokenExpiration:      time.Duration(config.Token.JWTExpiration) * time.Minute,
				RefreshTokenExpiration: time.Duration(config.Token.RefreshExpiration) * time.Hour,
			}), a.lg))
	if err != nil {
		return fmt.Errorf("failed to create gRPC server: %s", err.Error())
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
	errs.Go(a.repo.Close)
	errs.Go(func() error { return a.srv.Close(ctx) })

	return errs.Wait()
}
