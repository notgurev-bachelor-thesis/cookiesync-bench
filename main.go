package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
	"log"
	"math/rand/v2"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	requests    = flag.Int("r", 0, "Number of requests (approximate)")
	connections = flag.Int("c", 10, "Number of connections")
	duration    = flag.Duration("d", time.Duration(0), "Duration of benchmark")
	url         = flag.String("url", "http://cl-hot1-1.moevideo.net:8080", "URL of targeted server")
	verbose     = flag.Bool("v", false, "Verbose mode")
	threads     = flag.Int("t", runtime.NumCPU(), "Number of concurrent threads per connection")
	limit       = flag.Int("l", 0, "Limit requests per second")
)

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	go func() {
		<-ctx.Done()
		time.Sleep(time.Second)
		log.Println("Failed to end benchmark gracefully, killing process...")
		os.Exit(0)
	}()

	if *requests == 0 && *duration == time.Duration(0) {
		log.Println("Error: must provide duration or number of requests")
		return
	}

	if *verbose {
		log.Println("Running in verbose mode")
		log.Println("Warning: verbose mode severely decreases performance")
	}

	log.Printf("Target URL: %s\n", *url)
	log.Printf("Connections: %d\n", *connections)
	log.Printf("Threads (goroutines) per connection: %d\n", *threads)
	log.Printf("Rate limit: %d r/s\n", *limit)

	limiter := rate.NewLimiter(rate.Limit(*limit), 1) // no bursts

	start := time.Now()

	log.Printf("Starting benchmark at %s\n", start.Format(time.DateTime))

	var sent atomic.Int64

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
				elapsed := time.Since(start).Seconds()
				s := sent.Load()
				rps := float64(s) / elapsed
				log.Printf("Progress: sent = %d, avg rps = %.0f r/s, elapsed = %ds\n", s, rps, int(elapsed))
			}
		}
	}()

	var wg sync.WaitGroup
	if *requests == 0 {
		log.Printf("Duration: %s\n", duration.String())

		ctx, cancel = context.WithTimeout(ctx, *duration)
		defer cancel()

		for i := 0; i < *connections; i++ {
			client := &fasthttp.PipelineClient{
				Dial: func(addr string) (net.Conn, error) {
					return fasthttp.Dial(strings.TrimPrefix(*url, "http://"))
				},
				//MaxConns: math.MaxInt - 2000, // Ruins performance, for some reason
			}

			for t := 0; t < *threads; t++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := 0; ; j++ {
						select {
						case <-ctx.Done():
							return
						default:
							wait(ctx, limiter)
							send(client, j)
							sent.Add(1)
						}
					}
				}()
			}
		}
	} else {
		log.Printf("Sending requests: %d\n", *requests)

		req := make(chan int, *requests)
		for i := 0; i < *requests; i++ {
			req <- i
		}
		close(req)

		for i := 0; i < *connections; i++ {
			client := &fasthttp.PipelineClient{
				Dial: func(addr string) (net.Conn, error) {
					return fasthttp.Dial(strings.TrimPrefix(*url, "http://"))
				},
				//MaxConns: math.MaxInt - 2000, // Ruins performance, for some reason
			}

			for k := 0; k < *threads; k++ {
				wg.Add(1)
				go func() {
					defer wg.Done()

					for j := range req {
						wait(ctx, limiter)
						send(client, j)
						sent.Add(1)
					}
				}()
			}
		}
	}

	wg.Wait()

	log.Println("Benchmark finished")
	log.Printf("Sent total of %d requests\n", sent.Load())
	if *duration > 0 {
		log.Printf("Average RPS = %f\n", float64(sent.Load())/duration.Seconds())
	}
}

func wait(ctx context.Context, limiter *rate.Limiter) {
	if limiter != nil {
		_ = limiter.Wait(ctx)
	}
}

type Client interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

func send(client Client, i int) {
	uid := generateRandomUID(20)
	d := rand.IntN(20) + 1
	b := rand.IntN(1000000) + 1

	reqURL := fmt.Sprintf("%s?d=%d&b=%d", *url, d, b)

	req := fasthttp.AcquireRequest()

	req.SetRequestURI(reqURL)
	req.Header.SetCookie("uid", uid)

	if err := client.Do(req, nil); err != nil {
		log.Printf("Error: %s\n", err)
		return
	}

	fasthttp.ReleaseRequest(req)

	if *verbose {
		log.Printf("Request %d: UID %s, d %d, b %d\n", i+1, uid, d, b)
	}
}

func generateRandomUID(length int) string {
	const charset = "0123456789abcdef"
	b := strings.Builder{}
	for i := 0; i < length; i++ {
		b.WriteByte(charset[rand.IntN(len(charset))])
	}
	return b.String()
}
