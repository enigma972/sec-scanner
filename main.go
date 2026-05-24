package main

// Note: run this package with `go run .` (or `go run *.go`) so all files
// in the package are compiled. Running `go run main.go` builds only that
// single file and will report undefined references to handlers defined
// in other files.

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/api/scan", scanHandler)
	http.HandleFunc("/api/fuzzing", fuzzHandler)

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
