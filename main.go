package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

// X12Document and Segment structs remain the same...
type X12Document struct {
	Segments []Segment `json:"segments"`
}
type Segment struct {
	ID       string   `json:"id"`
	Elements []string `json:"elements"`
}

// ... parseX12 function remains the same ...
func parseX12(content string) (X12Document, error) {
	segmentTerminator := "~"
	elementSeparator := "*"
	rawSegments := strings.Split(content, segmentTerminator)
	doc := X12Document{}
	for _, rawSeg := range rawSegments {
		cleanSeg := strings.TrimSpace(rawSeg)
		if cleanSeg == "" {
			continue
		}
		elements := strings.Split(cleanSeg, elementSeparator)
		segmentID := elements[0]
		dataElements := []string{}
		if len(elements) > 1 {
			dataElements = elements[1:]
		}
		doc.Segments = append(doc.Segments, Segment{ID: segmentID, Elements: dataElements})
	}
	return doc, nil
}

// --- NEW: THE DOORMAN (Middleware) ---
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Get the Key from the Cloud Environment
		requiredKey := os.Getenv("API_SECRET")
		if requiredKey == "" {
			// Safety net: If you forget to set the key, default to "password" so it's not open
			requiredKey = "password"
		}

		// 2. Check the User's Key
		userKey := r.Header.Get("X-API-KEY")

		if userKey != requiredKey {
			http.Error(w, "Unauthorized: Invalid or missing API Key", http.StatusUnauthorized)
			return
		}

		// 3. Allow them through
		next(w, r)
	}
}

// The Handler (Business Logic)
func x12Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	parsedDoc, err := parseX12(string(bodyBytes))
	if err != nil {
		http.Error(w, "Failed to parse X12", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parsedDoc)
}

func main() {
	// WRAP the handler with the middleware
	http.HandleFunc("/convert", authMiddleware(x12Handler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := "0.0.0.0:" + port

	fmt.Printf("🔒 Secure X12 API running on %s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
