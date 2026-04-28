package main

import (
	"fmt"
	"io"
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
		fmt.Println(res.URL, res.StatusCode, res.BodyLength, string(res.Body), res.Err)
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

	fetchAsync(urls)

	// fetchSync(urls)

}
