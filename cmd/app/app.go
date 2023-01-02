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
	fmt.Printf("URLS: %+v", args)
	p := parser.NewParser(parser.SliceSource(args))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	visitor.Run(ctx, 1, p)
	fmt.Printf("\napplication is done")
}
