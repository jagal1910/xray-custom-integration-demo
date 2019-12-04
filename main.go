package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const apiKey = "secretKey"
const providerName = "custom-integration-demo"

type CheckAuthResponse struct {
	Valid bool
	Error string
}

type Component struct {
	ComponentID string `json:"component_id"`
	Blobs       []string
}

// Component data provided by XRay. We can use this to look up components.
type ComponentInfoRequest struct {
	Components []Component
	Context    string
}

type Vulnerability struct {
	CVE         string
	Type        string
	SourceID    string `json:"source_id"`
	Summary     string
	Description string
	CVSSV2      string `json:"cvss_v2"`
	URL         string
	PublishDate string `json:"publish_date"`
	References  []string
}

type ComponentInfo struct {
	ComponentID     string `json:"component_id"`
	Licenses        []string
	Provider        string // This should always be the name of your provider
	Vulnerabilities []Vulnerability
}

// XRay uses this info to check for violations
type ComponentInfoResponse struct {
	Components []ComponentInfo
}

func main() {
	// Routes: must be supplied to x-ray during integration setup as TestURL and URL
	http.HandleFunc("/api/checkauth", checkAuth) // TestURL
	http.HandleFunc("/api/componentinfo", componentInfo) // URL

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Something went wrong: ", err)
		return
	}
}

// The Test URL you can use to test your API key with the provider using the "Test" button
func checkAuth(w http.ResponseWriter, r *http.Request) {
	resp := CheckAuthResponse{true, ""}
	key := r.Header.Get("apiKey")
	if key != apiKey {
		resp = CheckAuthResponse{Valid: false, Error: "Invalid Api Key"}
	}
	js, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// This endpoint provides information to XRay about components
func componentInfo(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	requestPayload := ComponentInfoRequest{}
	err = json.Unmarshal(body, &requestPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	// Get all the components from the "db".
	// The db is just a json file with fake data about components.
	db := getDB()

	// Get matching components from db
	responsePayload := findComponents(requestPayload.Components, db)

	js, err := json.Marshal(responsePayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Read json file "database" into memory
func getDB() []ComponentInfo {
	file, _ := ioutil.ReadFile("db.json")
	var data []ComponentInfo
	_ = json.Unmarshal(file, &data)
	return data
}

// Search db for matching components and return
func findComponents(components []Component, db []ComponentInfo) ComponentInfoResponse {
	matches := ComponentInfoResponse{}
	// Check database for matching components
	for _, component := range components {
		result := ComponentInfo{}
		for _, item := range db {
			if item.ComponentID == component.ComponentID {
				result = item
				result.Provider = providerName
				break
			}
		}
		matches.Components = append(matches.Components, result)
	}
	return matches
}
