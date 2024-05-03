package main

import (
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"sync"
)

var (
	requests    = flag.Int("r", 1000, "Number of requests per connection")
	connections = flag.Int("c", 10, "Number of connections")
	url         = flag.String("url", "http://cl-hot1-1.moevideo.net:8080", "URL of targeted server")
)

func main() {
	flag.Parse()

	var wg sync.WaitGroup

	for i := 0; i < *connections; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < *requests; j++ {
				uid := generateRandomUID(20)
				d := rand.Intn(20) + 1
				b := rand.Intn(1000000) + 1

				reqURL := fmt.Sprintf("%s?d=%d&b=%d", *url, d, b)

				req := fasthttp.AcquireRequest()

				req.SetRequestURI(reqURL)
				req.Header.SetCookie("uid", uid)

				resp := fasthttp.AcquireResponse()

				if err := fasthttp.Do(req, resp); err != nil {
					fmt.Printf("Error: %s\n", err)
					return
				}

				fasthttp.ReleaseRequest(req)
				fasthttp.ReleaseResponse(resp)

				fmt.Printf("Request %d: Status %d, UID %s, d %d, b %d\n", j+1, resp.StatusCode(), uid, d, b)
			}
		}()
	}

	wg.Wait()
}

func generateRandomUID(length int) string {
	const charset = "0123456789abcdef"
	uid := make([]byte, length)
	for i := range uid {
		uid[i] = charset[rand.Intn(len(charset))]
	}
	return string(uid)
}
