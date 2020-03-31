package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joshturge-io/auth/pkg/cmd"
)

func main() {
	var (
		app = &cmd.App{}
		err error
	)
	if err = app.Initialise(os.Getenv("JWT_SECRET"), os.Getenv("REDIS_ADDR"),
		os.Getenv("GRPC_ADDR")); err != nil {
		log.Fatalf("ERROR: Initialisation: %s\n", err.Error())
	}
	if err = app.Start(); err != nil {
		log.Printf("ERROR: Starting: %s", err.Error())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err = app.Shutdown(ctx); err != nil {
		log.Fatalf("ERROR: Closing: %s", err.Error())
	}
}
