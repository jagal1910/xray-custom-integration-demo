package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

const providerName = "custom-integration-demo"
const InvalidAPIKeyMessage = "Invalid Api Key"

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
	Version     string
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
	dbPath, apiKey, err := parseArgs()
	if err != nil {
		fmt.Println("Something went wrong: ", err)
		return
	}
	router := CreateRouter(dbPath, apiKey)
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("Something went wrong: ", err)
		return
	}
}

// Only one api key is supported here, passed in as a CLI argument
func CreateRouter(dbPath string, apiKey string) *http.ServeMux {
	router := http.NewServeMux()
	// Routes: must be supplied to x-ray during integration setup as TestURL and URL
	router.HandleFunc("/api/checkauth", func(w http.ResponseWriter, r *http.Request) {
		checkAuth(w, r, apiKey)
	}) // TestURL
	router.HandleFunc("/api/componentinfo", func(w http.ResponseWriter, r *http.Request) {
		componentInfo(w, r, dbPath, apiKey)
	}) // URL
	return router
}

func parseArgs() (string, string, error) {
	if len(os.Args) < 2 {
		fmt.Println()
		return "", "", fmt.Errorf("\nApi key is required\nUsage: go run main.go (api-key)")
	}
	apiKey := os.Args[1]
	dbPath := "db.json"
	if len(os.Args) > 2 {
		dbPath = os.Args[2]
	}
	return dbPath, apiKey, nil
}

// The Test URL you can use to test your API key with the provider using the "Test" button
func checkAuth(w http.ResponseWriter, r *http.Request, apiKey string) {
	resp := CheckAuthResponse{true, ""}
	key := r.Header.Get("apiKey")
	if key != apiKey {
		resp = CheckAuthResponse{Valid: false, Error: InvalidAPIKeyMessage}
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
func componentInfo(w http.ResponseWriter, r *http.Request, dbPath string, apiKey string) {
	key := r.Header.Get("apiKey")
	if key != apiKey {
		http.Error(w, InvalidAPIKeyMessage, http.StatusUnauthorized)
	}
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
	db, err := getDB(dbPath)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get matching components from db
	responsePayload, err := findComponents(requestPayload.Components, db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	js, err := json.Marshal(responsePayload)
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Unmarshall data from the json db
func getDB(dbPath string) ([]ComponentInfo, error) {
	file, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}
	var data []ComponentInfo
	_ = json.Unmarshal(file, &data)
	return data, nil
}

// Search db for matching components and return
func findComponents(components []Component, db []ComponentInfo) (ComponentInfoResponse, error) {
	matches := ComponentInfoResponse{}
	// Check database for matching components
	for _, component := range components {
		result := ComponentInfo{}
		name, version := getVersionAndNameFromComponentID(component.ComponentID)
		for _, item := range db {
			if item.ComponentID == name {
				for _, vuln := range item.Vulnerabilities {
					isMatching, err := isVersionMatching(version, vuln.Version)
					if err != nil {
						return matches, err
					}
					if isMatching {
						result = item
						result.Provider = providerName
						// Restore the full component id to include the version so XRay can identify it.
						result.ComponentID = component.ComponentID
						break
					}
				}
			}
		}
		matches.Components = append(matches.Components, result)
	}
	return matches, nil
}

// Extract the version from the last ":" in the component ID
func getVersionAndNameFromComponentID(componentID string) (string, string) {
	index := strings.LastIndex(componentID, ":")
	split := strings.SplitAfterN(componentID, ":", index)
	return componentID[0:index], split[len(split)-1]
}

// Only semver is supported
func isVersionMatching(componentVersion string, versionRange string) (bool, error) {
	constraint, err := semver.NewConstraint(versionRange)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	candidate, err := semver.NewVersion(componentVersion)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return constraint.Check(candidate), nil
}
