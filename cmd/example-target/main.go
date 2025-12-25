package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

var (
	requestCount uint64
	errorCount   uint64
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	port := flag.Int("port", 8080, "Port to expose metrics on")
	flag.Parse()

	// Simulate some metrics changing
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			atomic.AddUint64(&requestCount, uint64(rand.Intn(10)+1))
			if rand.Float64() < 0.2 {
				atomic.AddUint64(&errorCount, 1)
			}
		}
	}()

	http.HandleFunc("/metrics", handleMetrics)
	http.HandleFunc("/", handleIndex)

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("Example target service running on %s\n", addr)
	fmt.Printf("Metrics available at http://localhost%s/metrics\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		panic(err)
	}
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>Example Target</title>
</head>
<body>
    <h1>Example Target Service</h1>
    <p>This is a sample service that exposes metrics for Promenitheus to scrape.</p>
    <p><a href="/metrics">View Metrics</a></p>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")

	// Write metrics in Prometheus format
	fmt.Fprintf(w, "# HELP http_requests_total Total number of HTTP requests\n")
	fmt.Fprintf(w, "# TYPE http_requests_total counter\n")
	fmt.Fprintf(w, "http_requests_total{method=\"GET\",endpoint=\"/api\"} %d\n", atomic.LoadUint64(&requestCount))
	fmt.Fprintf(w, "http_requests_total{method=\"POST\",endpoint=\"/api\"} %d\n", atomic.LoadUint64(&requestCount)/2)

	fmt.Fprintf(w, "# HELP http_errors_total Total number of HTTP errors\n")
	fmt.Fprintf(w, "# TYPE http_errors_total counter\n")
	fmt.Fprintf(w, "http_errors_total %d\n", atomic.LoadUint64(&errorCount))

	fmt.Fprintf(w, "# HELP memory_usage_bytes Current memory usage in bytes\n")
	fmt.Fprintf(w, "# TYPE memory_usage_bytes gauge\n")
	fmt.Fprintf(w, "memory_usage_bytes %d\n", rand.Int63n(1000000000)+500000000)

	fmt.Fprintf(w, "# HELP cpu_usage_percent Current CPU usage percentage\n")
	fmt.Fprintf(w, "# TYPE cpu_usage_percent gauge\n")
	fmt.Fprintf(w, "cpu_usage_percent %.2f\n", rand.Float64()*100)

	fmt.Fprintf(w, "# HELP active_connections Number of active connections\n")
	fmt.Fprintf(w, "# TYPE active_connections gauge\n")
	fmt.Fprintf(w, "active_connections %d\n", rand.Intn(100)+10)
}
