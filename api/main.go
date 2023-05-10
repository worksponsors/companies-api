package main

import (
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

const gzURL = "https://data.worksponsors.co.uk/master.json.gz"

type Company struct {
	Name   string  `json:"name"`
	Rating float64 `json:"rating"`
}

var jsonData []Company
var once sync.Once

func fetchAndDecompressJSON() {
	response, err := http.Get(gzURL)
	if err != nil {
		log.Fatalf("Error fetching the JSON file: %v", err)
	}

	defer response.Body.Close()

	compressedData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Error reading compressed data: %v", err)
	}

	decompressedData, err := zlib.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		log.Fatalf("Error decompressing data: %v", err)
	}

	jsonBytes, err := ioutil.ReadAll(decompressedData)
	if err != nil {
		log.Fatalf("Error reading JSON data: %v", err)
	}

	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		log.Fatalf("Error unmarshaling JSON data: %v", err)
	}
}

func searchByName(searchKey string) []Company {
	var searchResults []Company
	regexPattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(searchKey))

	for _, item := range jsonData {
		if matched, _ := regexp.MatchString(regexPattern, strings.ToLower(item.Name)); matched {
			searchResults = append(searchResults, item)
		}
	}

	return searchResults
}

func getCompaniesHandler(w http.ResponseWriter, r *http.Request) {
	companyNamesParam := r.URL.Query().Get("companyNames")
	if companyNamesParam == "" {
		http.Error(w, "Missing companyNames parameter", http.StatusBadRequest)
		return
	}

	companyNames := strings.Split(companyNamesParam, ",")

	results := make(map[string]interface{})
	for _, name := range companyNames {
		searchResults := searchByName(name)

		if len(searchResults) == 1 {
			results[name] = map[string]interface{}{
				"key":         name,
				"count":       len(searchResults),
				"exact_match": searchResults[0].Name,
				"exact_rating": searchResults[0].Rating,
			}
		} else {
			results[name] = map[string]interface{}{
				"key":          name,
				"count":        len(searchResults),
				"exact_match":  nil,
				"exact_rating": nil,
			}
		}
	}

	jsonResponse, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func main() {
	once.Do(fetchAndDecompressJSON)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/.netlify/functions/api", getCompaniesHandler)

	log.Printf("Server listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
