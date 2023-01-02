package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"
)

func main() {
	args := os.Args[1:]
	fmt.Printf("URLS: %+v", args)
	urlCh := parse(args)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	run(ctx, 1, urlCh)
	fmt.Printf("\napplication is done")
}

func run(ctx context.Context, concurrency int, urlCh <-chan string) {
	responseCh := make(chan response)
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for u := range urlCh {
				if err := ctx.Err(); err != nil {
					fmt.Printf("\nrun must stop: %v", ctx.Err())
					return
				}

				if err := visitURL(ctx, u, responseCh); err != nil {
					fmt.Printf("\ncould not visit url [%s]: %s", u, err)
					continue
				}
			}

			fmt.Printf("\nall urls have been visited")
		}()
	}

	doneCh := handleResponses(ctx, responseCh)

	wg.Wait()
	close(responseCh)
	<-doneCh
}

func handleResponses(ctx context.Context, responseCh chan response) <-chan struct{} {
	doneCh := make(chan struct{})

	go func() {
		defer close(doneCh)

		allResponses, err := sinkResponses(ctx, responseCh)
		if err != nil {
			return // todo: log
		}

		sort.Sort(allResponses)

		for i := range allResponses {
			fmt.Printf("\nURL: %s => BodySize: %d", allResponses[i].URL, allResponses[i].BodySize)
		}
	}()

	return doneCh
}

func sinkResponses(ctx context.Context, inCh chan response) (responses, error) {
	allResponses := make(responses, 0)
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("\n\nhandleResponses must stop: %v", ctx.Err())
			return nil, ctx.Err()
		case resp, ok := <-inCh:
			if ok {
				fmt.Printf("\nreceived response %s", resp.URL)
				allResponses = append(allResponses, resp)
			} else {
				return allResponses, nil
			}
		}
	}
}

type response struct {
	Method     string
	URL        string
	BodySize   int
	StatusCode int
}

type responses []response

func (r responses) Len() int {
	return len(r)
}

func (r responses) Less(i, j int) bool {
	return r[i].BodySize > r[j].BodySize // descending order
}

func (r responses) Swap(i, j int) {
	r[i], r[j] = r[j], r[i]
}

func visitURL(baseCtx context.Context, url string, resultCh chan<- response) error {
	ctx, cancel := context.WithTimeout(baseCtx, 2*time.Second)
	defer cancel()

	req, err := prepareRequest(ctx, url)
	if err != nil {
		fmt.Printf("could not build request for url [%s]", url)
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("\nrequest %s to url [%s] failed: %s", req.Method, url, err.Error())
		return err
	}

	result := response{
		Method:     req.Method,
		URL:        url,
		StatusCode: resp.StatusCode,
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("could not close body of url [%s] resp: %s", url, err.Error())
		}

		resultCh <- result
	}()

	if resp.StatusCode >= 300 {
		result.BodySize = 0
	} else {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("could not get body from url [%s]: %s", url, err.Error())
			return err
		}
		result.BodySize = len(body)
	}

	return nil
}

func prepareRequest(ctx context.Context, url string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "GO test exercise")
	return req, nil
}

func parse(urls []string) <-chan string {
	resultCh := make(chan string)

	go func() {
		defer close(resultCh)

		for _, s := range urls {
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