package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	requests    = flag.Int("r", 0, "Number of requests per connection")
	connections = flag.Int("c", 10, "Number of connections")
	duration    = flag.Duration("d", time.Duration(0), "Duration of benchmark")
	url         = flag.String("url", "http://cl-hot1-1.moevideo.net:8080", "URL of targeted server")
	verbose     = flag.Bool("v", false, "Verbose mode")
	threads     = flag.Int("t", runtime.NumCPU(), "Number of concurrent threads per connection")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		<-ctx.Done()
		time.Sleep(time.Second)
		fmt.Println("Failed to end benchmark gracefully, killing process...")
		os.Exit(0)
	}()

	if *requests == 0 && *duration == time.Duration(0) {
		fmt.Println("Error: must provide duration or number of requests")
		return
	}

	if *verbose {
		fmt.Println("Running in verbose mode")
		fmt.Println("Warning: verbose mode severely decreases performance")
	}

	fmt.Printf("Starting benchmark at %s\n", time.Now().Format(time.RFC850))
	fmt.Printf("Target URL: %s\n", *url)
	fmt.Printf("Connections: %d\n", *connections)
	fmt.Printf("Threads (goroutines) per connection: %d\n", *threads)

	var sent atomic.Int64
	var wg sync.WaitGroup
	if *requests == 0 {
		fmt.Printf("Duration: %s\n", duration.String())

		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()

		for i := 0; i < *connections; i++ {
			for t := 0; t < *threads; t++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					j := 0
					for {
						select {
						case <-ctx.Done():
							return
						default:
							j++
							send(j)
							sent.Add(1)
						}
					}
				}()
			}
		}
	} else {
		fmt.Printf("Number of requests: %d\n", *requests)
		for i := 0; i < *connections; i++ {
			for k := 0; k < *threads; k++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; j < *requests; j++ {
						send(j)
						sent.Add(1)
					}
				}()
			}
		}
	}

	wg.Wait()

	fmt.Println("Benchmark finished")
	fmt.Printf("Sent total of %d requests\n", sent.Load())
	if *duration > 0 {
		fmt.Printf("Average RPS = %f\n", float64(sent.Load())/duration.Seconds())
	}
}

func send(i int) {
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

	if *verbose {
		fmt.Printf("Request %d: Status %d, UID %s, d %d, b %d\n", i+1, resp.StatusCode(), uid, d, b)
	}
}

func generateRandomUID(length int) string {
	const charset = "0123456789abcdef"
	uid := make([]byte, length)
	for i := range uid {
		uid[i] = charset[rand.Intn(len(charset))]
	}
	return string(uid)
}
