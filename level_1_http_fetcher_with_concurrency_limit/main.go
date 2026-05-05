package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

type result struct {
	URL        string
	StatusCode int
	BodyLength int
	Body       []byte
	Err        error
}

func syncFetch(url string) (*result, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	result := &result{
		URL:        url,
		StatusCode: resp.StatusCode,
		BodyLength: int(resp.ContentLength),
		Err:        err,
	}
	return result, nil
}

func asyncFetch(url string, ch chan<- result) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		ch <- result{URL: url, Err: err}
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- result{URL: url, Err: err}
		return
	}
	ch <- result{
		URL:        url,
		StatusCode: resp.StatusCode,
		BodyLength: int(resp.ContentLength),
		Body:       body,
		Err:        nil,
	}
}

func asyncFetchWithCancelation(ctx context.Context, url string, ch chan<- result) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- result{URL: url, Err: err}
		return
	}
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	resChan := make(chan result, 1)
	go func() {
		resp, err := client.Do(req)
		if err != nil {
			resChan <- result{URL: url, Err: err}
			return
		}
		defer resp.Body.Close()

		resChan <- result{
			URL:        url,
			StatusCode: resp.StatusCode,
			BodyLength: int(resp.ContentLength),
			Err:        nil,
		}
	}()

	select {
	case res := <-resChan:
		ch <- res
	case <-ctx.Done():
		ch <- result{URL: url, Err: ctx.Err()}
	case <-time.After(2 * time.Second):
		ch <- result{URL: url, Err: fmt.Errorf("timed out")}
	}
}

func fetchSync(urls []string) {
	for _, url := range urls {
		resp, err := syncFetch(url)
		if err != nil {
			fmt.Println("Fehler:", err)
			return
		}
		fmt.Println(resp)
	}
}

func fetchAsync(urls []string) {
	ch := make(chan result)
	for _, url := range urls {
		go asyncFetch(url, ch)
	}
	for range urls {
		res := <-ch
		fmt.Println(res.URL, res.StatusCode, res.BodyLength, res.Err)
	}
}

func fetchAsyncWithCancelation(urls []string) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch := make(chan result)

	for _, url := range urls {
		go asyncFetchWithCancelation(ctx, url, ch)
	}

	for range urls {
		res := <-ch
		// if res.StatusCode >= 500 {
		// 	cancel()
		// }
		fmt.Println(res.URL, res.StatusCode, res.BodyLength, res.Err)
	}
}

func fetchWithRetry(maxRetries int, url string) (*result, error) {
	var lastErr error
	for i := range maxRetries {
		attempt := i
		res, err := syncFetch(url)

		if err == nil {
			return res, nil
		}

		lastErr = err

		if !isRetryable(err) {
			return res, err
		}

		if attempt <= maxRetries {
			backoff := time.Duration(attempt+1) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}
	return nil, fmt.Errorf("Failed after  %d retries: %w", maxRetries, lastErr)
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	if _, ok := errors.AsType[net.Error](err); ok {
		return true
	}
	return false
}

func fetchWithRetrySync(urls []string) {
	for _, url := range urls {
		res, err := fetchWithRetry(3, url)
		if err != nil {
			fmt.Println("Fehler:", err)
		} else {
			fmt.Println(res)
		}
	}
}

func main() {
	urls := []string{
		"https://httpbin.org/delay/1",
		"https://httpbin.org/delay/2",
		"https://httpbin.org/status/200",
		"https://httpbin.org/status/500",
		"https://httpbin.org/bytes/100",
		"https://httpbin.org/bytes/200",
		"https://httpbin.org/delay/3",
		"https://httpbin.org/status/404",
		"https://httpbin.org/bytes/50",
		"https://httpbin.org/delay/1",
	}

	fetchWithRetrySync(urls)
	// fetchAsyncWithCancelation(urls)

	// fetchSync(urls)

}
