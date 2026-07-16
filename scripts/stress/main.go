package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	url := os.Getenv("TARGET_URL")
	if url == "" {
		url = "https://api.gogate.rejwanul.dev/api/v1/users"
	}
	apiKey := os.Getenv("TEST_API_KEY")
	if apiKey == "" {
		log.Fatal("TEST_API_KEY environment variable is required")
	}

	concurrency := 50
	duration := 15 * time.Second

	fmt.Printf("Starting stress test on %s\n", url)
	fmt.Printf("Concurrency: %d, Duration: %s\n", concurrency, duration)
	fmt.Println("Please wait...")

	var (
		totalRequests int64
		success200    int64
		rateLimit429  int64
		otherStatus   int64
		errorsCount   int64
	)

	var wg sync.WaitGroup
	wg.Add(concurrency)

	startTime := time.Now()
	endTime := startTime.Add(duration)

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			client := &http.Client{Timeout: 5 * time.Second}
			for time.Now().Before(endTime) {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					atomic.AddInt64(&errorsCount, 1)
					continue
				}
				req.Header.Set("X-API-Key", apiKey)

				atomic.AddInt64(&totalRequests, 1)
				resp, err := client.Do(req)
				if err != nil {
					atomic.AddInt64(&errorsCount, 1)
					continue
				}

				if resp.StatusCode == 200 {
					atomic.AddInt64(&success200, 1)
				} else if resp.StatusCode == 429 {
					atomic.AddInt64(&rateLimit429, 1)
				} else {
					atomic.AddInt64(&otherStatus, 1)
				}
				resp.Body.Close()
				
				// Small sleep to control the firehose slightly
				time.Sleep(10 * time.Millisecond) 
			}
		}()
	}

	wg.Wait()

	fmt.Println("\n--- Stress Test Results ---")
	fmt.Printf("Total Requests Sent: %d\n", totalRequests)
	fmt.Printf("Successful (200 OK): %d\n", success200)
	fmt.Printf("Rate Limited (429):  %d\n", rateLimit429)
	fmt.Printf("Other Status Codes:  %d\n", otherStatus)
	fmt.Printf("Connection Errors:   %d\n", errorsCount)
	fmt.Printf("Req/Sec Achieved:    %.2f\n", float64(totalRequests)/duration.Seconds())
}
