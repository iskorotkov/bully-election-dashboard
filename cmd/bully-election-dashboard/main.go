package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iskorotkov/bully-election-dashboard/pkg/collect"
	"github.com/iskorotkov/bully-election-dashboard/pkg/state"
	"github.com/iskorotkov/bully-election-dashboard/pkg/ui"
	"github.com/rs/cors"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
)

var (
	sleepTimeout = time.Second
)

func main() {
	var (
		logger *zap.Logger
		err    error
	)
	if os.Getenv("DEVELOPMENT") != "" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		log.Fatalf("couldn't create logger: %v", err)
	}

	defer logger.Sync()

	defer func() {
		if p := recover(); p != nil {
			logger.Fatal("panic occurred",
				zap.Any("panic", p))
		}
	}()

	namespace := os.Getenv("TARGET_NAMESPACE")
	if namespace == "" {
		logger.Fatal("namespace wasn't set")
	}

	stateServer := state.NewServer(logger.Named("state-server"))

	uiServer, err := ui.NewServer(namespace, logger.Named("ui-server"))
	if err != nil {
		logger.Fatal("couldn't create ui server",
			zap.Error(err))
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", uiServer.Handle)
	mux.HandleFunc("/api", stateServer.Handle)

	server := http.Server{
		Addr:    ":80",
		Handler: cors.Default().Handler(mux),
	}

	collector, err := collect.NewCollector(namespace, time.Second*5, logger.Named("collector"))
	if err != nil {
		logger.Fatal("couldn't create collector",
			zap.String("namespace", namespace),
			zap.Error(err))
	}

	go func() {
		for {
			data, err := collector.Collect()
			if err != nil {
				logger.Error("couldn't collect data",
					zap.Error(err))
			} else {
				stateServer.Update(data)
			}

			time.Sleep(sleepTimeout)
		}
	}()

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed",
				zap.Error(err))
		}
	}()

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("server shutdown failed",
			zap.Error(err))
	}
}
