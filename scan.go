package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

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
