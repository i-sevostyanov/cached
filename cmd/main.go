package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"probe/internal/cache"
	"probe/internal/server"
)

var errCanceled = errors.New("canceled")

func main() {
	address := flag.String("addr", "127.0.0.1:4000", "TCP server address")
	dataFile := flag.String("path", "data.bin", "Path to data file")
	interval := flag.Duration("interval", time.Minute, "Eviction interval")
	flag.Parse()

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)

	file, err := os.OpenFile(*dataFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		logger.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		if err = file.Close(); err != nil {
			logger.Printf("Failed to close file: %v", err)
		}
	}()

	inMemCache := cache.New()

	if err = inMemCache.Restore(file); err != nil {
		logger.Printf("Warning. Failed to restore cache: %v", err)
	}

	gr, ctx := errgroup.WithContext(context.Background())

	gr.Go(func() error {
		return server.New(*address, inMemCache, logger).Listen(ctx)
	})

	gr.Go(func() error {
		inMemCache.Eviction(ctx, *interval)
		return nil
	})

	gr.Go(func() error {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-signals:
				logger.Println("Caught stop signal. Exiting...")
				return errCanceled
			}
		}
	})

	if err = gr.Wait(); err != nil && err != errCanceled {
		logger.Fatalf("Failed to start: %v", err)
	}

	if err := file.Truncate(0); err != nil {
		logger.Fatalf("Failed to truncate file: %v", err)
	}

	if err := inMemCache.Dump(file); err != nil {
		logger.Fatalf("Failed to dump cache: %v", err)
	}
}
