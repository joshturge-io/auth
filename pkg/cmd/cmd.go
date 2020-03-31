package cmd

import (
	"context"
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
func (a *App) Initialise(jwtSecret, dbAddr, grpcAddr string) (err error) {
	a.lg = log.New(os.Stdout, "[INFO] ", log.Ltime|log.Ldate)

	a.lg.Println("Creating connection to database")

	a.repo, err = repository.NewRedisRepository(dbAddr)
	if err != nil {
		return fmt.Errorf("failed to make connection to database: %s", err.Error())
	}

	a.lg.Printf("Creating gRPC server on: %s\nRegistering gRPC services\n", grpcAddr)

	a.srv, err = grpc.NewServer(grpcAddr, service.NewGRPCAuthService(auth.NewService(jwtSecret, a.repo,
		// TODO(joshturge): pass from config file
		&auth.Options{RefreshTokenLength: 32,
			JWTokenExpiration:      15 * time.Minute,
			RefreshTokenExpiration: 24 * time.Hour}), a.lg))
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
