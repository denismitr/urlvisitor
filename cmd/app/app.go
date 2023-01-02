package main

import (
	"context"
	"fmt"
	"github.com/denismitr/urlvisitor/internal/parser"
	"github.com/denismitr/urlvisitor/internal/visitor"
	"os"
	"time"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	visitor.Run(ctx, 2, p)
	fmt.Printf("\napplication is done")
}
