package visitor

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

// makes visitor easier to test
type urlParser interface {
	URLs() <-chan string
}

type webClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type URLVisitor struct {
	// makes it testable
	client webClient
}

func NewDefaultClient(d time.Duration) webClient {
	return &http.Client{
		Timeout: d,
	}
}

func NewURLVisitor(client webClient) *URLVisitor {
	return &URLVisitor{client: client}
}

func (v *URLVisitor) Run(ctx context.Context, concurrency int, p urlParser) {
	responseCh := make(chan visitResult)

	var wg sync.WaitGroup
	urlsCh := p.URLs()
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for u := range urlsCh {
				if err := ctx.Err(); err != nil {
					log.Warn().Msgf("URLVisitor Run must stop: %v", ctx.Err())
					return
				}

				if err := v.visitURL(ctx, u, responseCh); err != nil {
					log.Error().Msgf("could not visit url [%s]: %s", u, err.Error())
					continue
				}
			}

		}()
	}

	doneCh := handleResponses(ctx, responseCh)
	wg.Wait()

	log.Info().Msgf("all urls have been visited")

	close(responseCh)
	<-doneCh
}

func handleResponses(ctx context.Context, responseCh chan visitResult) <-chan struct{} {
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		allVisits, err := sinkVisitResults(ctx, responseCh)
		if err != nil {
			log.Error().Msgf("could not retrieve all visitResults as a slice: %s", err.Error())
			return
		}

		sort.Sort(allVisits)

		for i := range allVisits {
			log.Info().Msgf("URL: %s => BodySize: %d", allVisits[i].URL, allVisits[i].BodySize)
		}
	}()

	return doneCh
}

func sinkVisitResults(ctx context.Context, visitCh chan visitResult) (visitResults, error) {
	allVisits := make(visitResults, 0)

	for {
		select {
		case <-ctx.Done():
			log.Warn().Msgf("sinkVisitResults must stop: %v", ctx.Err())
			return nil, ctx.Err()
		case visit, ok := <-visitCh:
			if ok {
				allVisits = append(allVisits, visit)
			} else {
				return allVisits, nil
			}
		}
	}
}

func (v *URLVisitor) visitURL(ctx context.Context, url string, resultCh chan<- visitResult) error {
	req, err := prepareRequest(ctx, url)
	if err != nil {
		return fmt.Errorf("could not build request for url [%s]: %w", url, err)
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return fmt.Errorf("request %s to url [%s] failed: %w", req.Method, url, err)
	}

	result := visitResult{
		Method:     req.Method,
		URL:        url,
		StatusCode: resp.StatusCode,
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Error().Msgf("could not close body of url [%s] resp: %s", url, err.Error())
		}
	}()

	if resp.StatusCode >= 300 {
		result.BodySize = 0
		log.Error().Msgf("received non 200 code %d from url [%s]", resp.StatusCode, url)
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().Msgf("could not get body from url [%s]: %v", url, err)
		}
		result.BodySize = len(body)
	}

	resultCh <- result
	return nil
}

func prepareRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("could not prepare request: %w", err)
	}
	req.Header.Set("User-Agent", "GO test exercise")
	return req, nil
}
