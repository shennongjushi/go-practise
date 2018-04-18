package main

import (
	"fmt"
	"sync"
)

type Cache struct {
	v 	map[string]bool
	mux sync.Mutex
}

var c = Cache{v: make(map[string]bool)}

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, done chan int) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	defer close(done)
	if depth <= 0 {
		done <- 1
		return
	}
	
	c.mux.Lock()
	if c.v[url] == false {
		c.v[url] = true
	} else {
		c.mux.Unlock()
		done <- 1
		return
	}
	c.mux.Unlock()
	
	body, urls, err := fetcher.Fetch(url)
	if err != nil {
		fmt.Println(err)
		done <- 1
		return
	}
	fmt.Printf("found: %s %q\n", url, body)
	
	childDone := make([]chan int, len(urls))
	for i, u := range urls {
		childDone[i] = make(chan int)
		go Crawl(u, depth-1, fetcher, childDone[i])
	}
	for i := range childDone {
		<- childDone[i]
	}
	done <- 1
	return
}

func main() {
	done := make(chan int)
	go Crawl("https://golang.org/", 4, fetcher, done)
	<- done
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"https://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"https://golang.org/pkg/",
			"https://golang.org/cmd/",
		},
	},
	"https://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"https://golang.org/",
			"https://golang.org/cmd/",
			"https://golang.org/pkg/fmt/",
			"https://golang.org/pkg/os/",
		},
	},
	"https://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
	"https://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"https://golang.org/",
			"https://golang.org/pkg/",
		},
	},
}

