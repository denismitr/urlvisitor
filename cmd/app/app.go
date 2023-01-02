package main

import (
	"context"
	"fmt"
	"github.com/denismitr/urlvisitor/internal/parser"
	"github.com/denismitr/urlvisitor/internal/visitor"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	args := os.Args[1:]
	var sourceFunc parser.SourceFunc
	if len(args) == 0 {
		// read from stdin - cat urls.txt | ./app
		sourceFunc = parser.ReaderSource(os.Stdin)
	} else {
		sourceFunc = parser.SliceSource(args)
	}

	p := parser.NewParser(sourceFunc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gracefulShutdown(cancel)

	visitor.Run(ctx, 2, p)
	fmt.Printf("\napplication is done!")
}

func gracefulShutdown(cancel context.CancelFunc) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(ch)
	<-ch
	cancel()
}
