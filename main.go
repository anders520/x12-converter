package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// X12Document represents the JSON structure
type X12Document struct {
	Segments []Segment `json:"segments"`
}

type Segment struct {
	ID       string   `json:"id"`
	Elements []string `json:"elements"`
}

// 1. The Core Logic (Refactored into a clean function)
func parseX12(content string) (X12Document, error) {
	// Standard X12 Delimiters
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

		doc.Segments = append(doc.Segments, Segment{
			ID:       segmentID,
			Elements: dataElements,
		})
	}
	return doc, nil
}

// 2. The API Handler
func x12Handler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}

	// Read the raw body (the X12 file)
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Convert bytes to string
	rawX12 := string(bodyBytes)

	// Run the parser
	parsedDoc, err := parseX12(rawX12)
	if err != nil {
		http.Error(w, "Failed to parse X12", http.StatusInternalServerError)
		return
	}

	// Return JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(parsedDoc)
}

func main() {
	// 3. Start the Server
	http.HandleFunc("/convert", x12Handler)

	port := ":8080"
	fmt.Printf("🚀 X12 API Server running on http://localhost%s\n", port)
	fmt.Println("   Waiting for POST requests...")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
