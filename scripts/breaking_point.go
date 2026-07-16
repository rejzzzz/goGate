package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

var (
	url         = flag.String("url", getEnvOrDefault("TARGET_URL", "https://api.gogate.rejwanul.dev/api/v1/users"), "Target URL to stress test")
	concurrency = flag.Int("c", 100, "Number of concurrent workers (goroutines)")
	duration    = flag.Duration("d", 30*time.Second, "Duration of the test")
	apiKey      = flag.String("apikey", "", "API Key for auth (defaults to TEST_API_KEY env var if empty)")
	bypassLimit = flag.Bool("bypass", true, "Spoof X-Forwarded-For to bypass per-IP rate limits")
)

func main() {
	flag.Parse()

	if *apiKey == "" {
		*apiKey = os.Getenv("TEST_API_KEY")
	}

	fmt.Printf("🚀 Starting Breaking Point Stress Test 🚀\n")
	fmt.Printf("URL: %s\n", *url)
	fmt.Printf("Concurrency: %d workers\n", *concurrency)
	fmt.Printf("Duration: %v\n", *duration)
	fmt.Println("---------------------------------------------------")

	// Create a custom HTTP client optimized for high throughput
	transport := &http.Transport{
		MaxIdleConns:        *concurrency * 2,
		MaxIdleConnsPerHost: *concurrency * 2,
		IdleConnTimeout:     30 * time.Second,
		DisableCompression:  true,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	var totalReqs uint64
	var totalSuccess uint64
	var totalFailed uint64
	var totalRatelimited uint64
	var totalGatewayErrors uint64

	var wg sync.WaitGroup
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	startTime := time.Now()

	for i := 0; i < *concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			// Worker local random generator to avoid lock contention
			rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

			for {
				select {
				case <-ctx.Done():
					return
				default:
					req, err := http.NewRequest("GET", *url, nil)
					if err != nil {
						continue
					}

					if *apiKey != "" {
						req.Header.Set("X-API-Key", *apiKey)
					}

					// Spoof IP to bypass per-ip rate limit and truly test proxy throughput
					if *bypassLimit {
						fakeIP := fmt.Sprintf("%d.%d.%d.%d", rnd.Intn(256), rnd.Intn(256), rnd.Intn(256), rnd.Intn(256))
						req.Header.Set("X-Forwarded-For", fakeIP)
					}

					resp, err := client.Do(req)
					atomic.AddUint64(&totalReqs, 1)

					if err != nil {
						atomic.AddUint64(&totalFailed, 1)
						continue
					}

					// Read and discard body to reuse the connection
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()

					switch resp.StatusCode {
					case 200, 201, 202, 204:
						atomic.AddUint64(&totalSuccess, 1)
					case 429:
						atomic.AddUint64(&totalRatelimited, 1)
					case 502, 503, 504:
						atomic.AddUint64(&totalGatewayErrors, 1)
						atomic.AddUint64(&totalFailed, 1)
					default:
						atomic.AddUint64(&totalFailed, 1)
					}
				}
			}
		}(i)
	}

	// Print live stats
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				elapsed := time.Since(startTime).Seconds()
				reqs := atomic.LoadUint64(&totalReqs)
				rps := float64(reqs) / elapsed
				fmt.Printf("[Live] RPS: %.0f | Reqs: %d | Success: %d | 429s: %d | 50x: %d\n", 
					rps, reqs, atomic.LoadUint64(&totalSuccess), atomic.LoadUint64(&totalRatelimited), atomic.LoadUint64(&totalGatewayErrors))
			}
		}
	}()

	wg.Wait()
	elapsed := time.Since(startTime)

	fmt.Println("\n🏁 Test Completed 🏁")
	fmt.Println("---------------------------------------------------")
	fmt.Printf("Total Time: %v\n", elapsed)
	fmt.Printf("Total Requests: %d\n", totalReqs)
	fmt.Printf("Throughput: %.2f req/sec\n", float64(totalReqs)/elapsed.Seconds())
	fmt.Printf("Success (2xx): %d\n", totalSuccess)
	fmt.Printf("Rate Limited (429): %d\n", totalRatelimited)
	fmt.Printf("Gateway Errors (50x): %d\n", totalGatewayErrors)
	fmt.Printf("Other Failures: %d\n", totalFailed-totalGatewayErrors)
}
