package main

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
