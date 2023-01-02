package parser

import (
	"fmt"
	"net/url"
	"strings"
)

type SourceFunc func() <-chan string

func SliceSource(urls []string) SourceFunc {
	return func() <-chan string {
		resultCh := make(chan string)
		go func() {
			defer close(resultCh)
			for _, u := range urls {
				resultCh <- strings.Trim(u, " ")
			}
		}()

		return resultCh
	}
}

type Parser struct {
	source SourceFunc
}

func NewParser(source SourceFunc) *Parser {
	return &Parser{source: source}
}

func (p *Parser) Parse() <-chan string {
	resultCh := make(chan string)

	go func() {
		defer close(resultCh)

		for s := range p.source() {
			if parsed, err := url.Parse(s); err != nil {
				fmt.Printf("\ninvalid url %s: %s", s, err.Error())
				continue
			} else {
				if parsed.Scheme == "" {
					parsed.Scheme = "http"
				}

				parsedURL := parsed.String()
				fmt.Printf("\nURL %s is ok", parsedURL)
				resultCh <- parsedURL
			}
		}

		fmt.Printf("\nAll URLs parsed!")
	}()

	return resultCh
}
