package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/joshturge-io/auth/pkg/cmd"
)

var configDir *string

func init() {
	configDir = flag.String("config", ".", "path to config dir")
}

func main() {
	flag.Parse()

	var (
		app = &cmd.App{}
		err error
	)

	if err = app.Initialise(*configDir); err != nil {
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
