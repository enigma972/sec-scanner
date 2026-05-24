package main

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
)

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
