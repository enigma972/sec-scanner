package main

import (
	"net/http"
	"time"
)

// Shared HTTP client for all outgoing requests (tuned timeout)
var httpClient = &http.Client{Timeout: 10 * time.Second}
