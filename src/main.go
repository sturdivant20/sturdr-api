package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sturdivant20/sturdr-api/include/api"
)

func main() {
	var app api.Application

	// Create a context that listens for SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// initialize server
	log.Println("Parsing TOML settings ...")
	if err := app.Init("./config/settings.toml"); err != nil {
		log.Fatalf("Failed to initialize database!")
	}

	// run server
	if err := app.Run(ctx, app.Mount()); err != nil {
		log.Fatalf("Failed to start server! %s", err.Error())
	}
}
