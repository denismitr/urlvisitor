package main

import (
	"context"
	"fmt"
	"github.com/denismitr/urlvisitor/internal/parser"
	"github.com/denismitr/urlvisitor/internal/visitor"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	defaultTimeout     = 2 * time.Second
	defaultConcurrency = 5
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	args := os.Args[1:]
	var sourceFunc parser.SourceFunc
	if len(args) == 0 {
		// read from stdin - cat urls.txt | ./app
		stdin, err := resolveStdin()
		if err != nil {
			log.Error().Msgf("input error: %v", err)
			os.Exit(1)
		}
		sourceFunc = parser.ReaderSource(ctx, stdin)
	} else {
		sourceFunc = parser.SliceSource(ctx, args)
	}

	p := parser.NewParser(sourceFunc)

	go gracefulShutdown(cancel)

	concurrency := resolveMaxConcurrency()
	log.Info().Msgf("running visitor with concurrency %d", concurrency)

	client := visitor.NewDefaultClient(resolveHttpTimeout())
	visitor.NewURLVisitor(client).Run(ctx, concurrency, p)

	log.Info().Msgf("application is done!")
}

func resolveStdin() (io.Reader, error) {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	size := fi.Size()
	if size == 0 {
		return nil, fmt.Errorf("no urls were provided via stdin")
	}

	return os.Stdin, nil
}

func resolveHttpTimeout() time.Duration {
	ts := os.Getenv("HTTP_TIMEOUT_SECONDS")
	if ts == "" {
		return defaultTimeout
	}

	seconds, err := strconv.Atoi(ts)
	if err != nil {
		return defaultTimeout
	}

	return time.Duration(seconds) * time.Second
}

func resolveMaxConcurrency() int {
	c := os.Getenv("MAX_CONCURRENCY")
	if c == "" {
		return defaultConcurrency
	}

	concurrency, err := strconv.Atoi(c)
	if err != nil {
		return defaultConcurrency
	}

	return concurrency
}

func gracefulShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
}
