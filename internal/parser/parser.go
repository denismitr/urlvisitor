package parser

import (
	"bufio"
	"github.com/rs/zerolog/log"
	"io"
	"net/url"
	"regexp"
	"strings"
)

// taken from stackoverflow - not tested well
var urlRegexp = regexp.MustCompile(`[(http(s)?):\/\/(www\.)?a-zA-Z0-9@:%._\+~#=]{2,256}\.[a-z]{2,6}\b([-a-zA-Z0-9@:%_\+.~#?&//=]*)`)

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

func ReaderSource(r io.Reader) SourceFunc {
	return func() <-chan string {
		resultCh := make(chan string)
		go func() {
			defer close(resultCh)
			scanner := bufio.NewScanner(r)
			scanner.Split(bufio.ScanLines)
			for scanner.Scan() {
				resultCh <- strings.Trim(scanner.Text(), " ")
			}
			return
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

func (p *Parser) URLs() <-chan string {
	resultCh := make(chan string)

	go func() {
		defer close(resultCh)

		for s := range p.source() {
			if !urlRegexp.MatchString(s) {
				log.Error().Msgf("invalid url [%s]", s)
				continue
			}

			if parsed, err := url.Parse(s); err != nil {
				log.Error().Msgf("invalid url [%s]: %s", s, err.Error())
				continue
			} else {
				if parsed.Scheme == "" {
					parsed.Scheme = "http"
				}

				parsedURL := parsed.String()
				resultCh <- parsedURL
			}
		}
	}()

	return resultCh
}
