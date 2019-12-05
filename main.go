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

type License struct {
	Version  string
	Licenses []string
}

// The shape of a component in the db used by this demo
type ComponentRecord struct {
	ComponentID     string `json:"component_id"`
	Licenses        []License
	Vulnerabilities []Vulnerability
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
		fmt.Println("\nSomething went wrong: ", err)
		return
	}
	router := CreateRouter(dbPath, apiKey)
	err = http.ListenAndServe(":8080", router)
	if err != nil {
		fmt.Println("\nSomething went wrong: ", err)
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
		return "", "", fmt.Errorf("Api key is required\nUsage: go run main.go (api-key) [path-to-db-file]")
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
		return
	}
	fmt.Println("we got past here")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	requestPayload := ComponentInfoRequest{}
	err = json.Unmarshal(body, &requestPayload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get all the components from the "db".
	// The db is just a json file with fake data about components.
	db, err := getDB(dbPath)
	if err != nil {
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Unmarshall data from the json db
func getDB(dbPath string) ([]ComponentRecord, error) {
	file, err := ioutil.ReadFile(dbPath)
	if err != nil {
		return nil, err
	}
	var data []ComponentRecord
	_ = json.Unmarshal(file, &data)
	return data, nil
}

// Search db for matching components and return
func findComponents(components []Component, db []ComponentRecord) (ComponentInfoResponse, error) {
	matches := ComponentInfoResponse{}
	// Check database for matching components
	for _, component := range components {
		result := ComponentInfo{}
		name, version := getVersionAndNameFromComponentID(component.ComponentID)
		for _, item := range db {
			if item.ComponentID == name {
				// Any Matching Licenses?
				licenses, err := getLicensesForVersion(version, item.Licenses)
				if err != nil {
					return matches, err
				}
				// Any Matching Vulnerabilities?
				vulnerabilities, err := getVulnerabilitiesForVersion(version, item.Vulnerabilities)
				if err != nil {
					return matches, err
				}
				if len(licenses) > 0 || len(vulnerabilities) > 0 {
					result = ComponentInfo{
						ComponentID:     component.ComponentID,
						Licenses:        licenses,
						Provider:        providerName,
						Vulnerabilities: vulnerabilities,
					}
					matches.Components = append(matches.Components, result)
					break
				}
			}
		}
	}
	return matches, nil
}

// Extract the version from the last ":" in the component ID
func getVersionAndNameFromComponentID(componentID string) (string, string) {
	index := strings.LastIndex(componentID, ":")
	name := ""
	version := ""
	if index > -1 {
		split := strings.SplitAfterN(componentID, ":", index)
		name = componentID[0:index]
		version = split[len(split)-1]
	}
	return name, version
}

func getLicensesForVersion(version string, licenses []License) ([]string, error) {
	var matchingLicences []string
	for _, license := range licenses {
		isMatching, err := isVersionMatching(version, license.Version)
		if err != nil {
			return matchingLicences, err
		}
		if isMatching {
			matchingLicences = append(matchingLicences, license.Licenses...)
		}
	}
	return matchingLicences, nil
}

func getVulnerabilitiesForVersion(version string, vulnerabilities []Vulnerability) ([]Vulnerability, error) {
	var matchingVulnerabilities []Vulnerability
	for _, vulnerability := range vulnerabilities {
		isMatching, err := isVersionMatching(version, vulnerability.Version)
		if err != nil {
			return matchingVulnerabilities, err
		}
		if isMatching {
			matchingVulnerabilities = append(matchingVulnerabilities, vulnerability)
		}
	}
	return matchingVulnerabilities, nil
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
