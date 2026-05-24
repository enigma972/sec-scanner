package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Shared HTTP client for all outgoing requests (tuned timeout)
var httpClient = &http.Client{Timeout: 10 * time.Second}

// ScanReport represents the result of scanning a single target
type ScanReport struct {
	Target string `json:"target_url"`
	Status int    `json:"status_code"`
	Online bool   `json:"is_online"`
}

// FuzzReport represents the result of sending a single fuzz payload
type FuzzReport struct {
	Payload    string `json:"payload"`
	Status     int    `json:"status_code"`
	BodyLength int    `json:"body_length"`
	Error      string `json:"error,omitempty"`
}

// scanTarget performs an HTTP GET on the provided target and sends a ScanReport
func scanTarget(target string, wg *sync.WaitGroup, results chan<- ScanReport) {
	defer wg.Done()

	resp, err := httpClient.Get(target)
	if err != nil {
		log.Printf("error connecting to %s: %v", target, err)
		results <- ScanReport{Target: target, Status: 0, Online: false}
		return
	}
	defer resp.Body.Close()

	online := resp.StatusCode < 400
	log.Printf("scanned %s -> status %d", target, resp.StatusCode)
	results <- ScanReport{Target: target, Status: resp.StatusCode, Online: online}
}

// scanHandler handles /api/scan?urls=... and returns a JSON array of ScanReport
func scanHandler(w http.ResponseWriter, r *http.Request) {
	// Query parameters are read-only; no body processing required
	query := r.URL.Query()
	if !query.Has("urls") {
		http.Error(w, "urls parameter missing", http.StatusBadRequest)
		return
	}

	targets := query["urls"]

	log.Println("starting multi-target scan")

	var wg sync.WaitGroup
	resultsCh := make(chan ScanReport, len(targets))

	for _, t := range targets {
		wg.Add(1)
		go scanTarget(t, &wg, resultsCh)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var reports []ScanReport
	for r := range resultsCh {
		reports = append(reports, r)
	}

	log.Println("multi-target scan complete")

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reports); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// fuzzHandler sends multiple payloads to a target via POST and returns results.
// Query parameters:
// - url (required): target to POST to
// - payloads (optional, multi): payload strings to send. If omitted, a small default set is used.
func fuzzHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	target := query.Get("url")
	if target == "" {
		http.Error(w, "url parameter missing", http.StatusBadRequest)
		return
	}

	payloads := query["payloads"]
	if len(payloads) == 0 {
		payloads = []string{"admin' --", "<script>alert(1)</script>", "../../etc/passwd", "\" OR 1=1 --"}
	}

	// Concurrency limiter to avoid overwhelming the target or local resources
	const maxConcurrency = 10
	limiter := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	resultsCh := make(chan FuzzReport, len(payloads))

	for _, p := range payloads {
		wg.Add(1)
		limiter <- struct{}{} // acquire
		go func(payload string) {
			defer wg.Done()
			defer func() { <-limiter }() // release

			resp, err := httpClient.Post(target, "text/plain", strings.NewReader(payload))
			if err != nil {
				resultsCh <- FuzzReport{Payload: payload, Status: 0, BodyLength: 0, Error: err.Error()}
				return
			}
			defer resp.Body.Close()

			body, _ := io.ReadAll(resp.Body)
			resultsCh <- FuzzReport{Payload: payload, Status: resp.StatusCode, BodyLength: len(body)}
		}(p)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	var reports []FuzzReport
	for r := range resultsCh {
		reports = append(reports, r)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(reports); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func main() {
	http.HandleFunc("/api/scan", scanHandler)
	http.HandleFunc("/api/fuzzing", fuzzHandler)

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
