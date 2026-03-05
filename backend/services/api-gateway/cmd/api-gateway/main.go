package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := loadRuntimeConfig()
	if err := validateRuntimeConfig(cfg); err != nil {
		log.Fatal(err)
	}

	startupCtx, cancelStartup := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelStartup()

	deps, cleanup, err := initDependencies(startupCtx, cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	server, serverCleanup := buildHTTPServer(cfg, deps)
	defer serverCleanup()
	runHTTPServer(server)
}

func runHTTPServer(server *http.Server) {
	stopCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-stopCtx.Done()
		shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancelShutdown()
		_ = server.Shutdown(shutdownCtx)
	}()

	log.Printf("api-gateway listening on %s", server.Addr)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
