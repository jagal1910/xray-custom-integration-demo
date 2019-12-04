package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

const apiKey = "secretKey"

func TestApi(t *testing.T) {
	// Create test server
	ts := httptest.NewServer(CreateRouter("db.json", apiKey))
	defer ts.Close()

	// CheckAuth endpoint
	t.Run("CheckAuth: Valid Api Key", func(t *testing.T) {
		validAPIKeyTest(t, ts)
	})
	t.Run("CheckAuth: Invalid Api Key", func(t *testing.T) {
		invalidAPIKeyTest(t, ts)
	})
	// ComponentInfo endpoint
	t.Run("ComponentInfo: Valid Api Key", func(t *testing.T) {
		validAPIKeyTestComponentInfo(t, ts)
	})
	t.Run("ComponentInfo: Invalid Api Key", func(t *testing.T) {
		invalidAPIKeyTestComponentInfo(t, ts)
	})

}

func validAPIKeyTest(t *testing.T, ts *httptest.Server) {
	req, err := http.NewRequest("GET", ts.URL+"/api/checkauth", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("apiKey", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data := CheckAuthResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatal(err)
	}
	if !data.Valid {
		t.Error("Failed to validate api key: ", apiKey)
	}
}

func invalidAPIKeyTest(t *testing.T, ts *httptest.Server) {
	req, err := http.NewRequest("GET", ts.URL+"/api/checkauth", nil)
	if err != nil {
		t.Fatal(err)
	}
	invalidKey := "invalidAPIKey"
	req.Header.Set("apiKey", invalidKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data := CheckAuthResponse{}
	body, err := ioutil.ReadAll(resp.Body)
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatal(err)
	}
	if data.Valid {
		t.Error("Invalid API key was accepted: ", invalidKey)
	}
	if data.Error != InvalidAPIKeyMessage {
		t.Error("Expected message: '"+InvalidAPIKeyMessage+"'. Got: '", data.Error, "'")
	}
}

func validAPIKeyTestComponentInfo(t *testing.T, ts *httptest.Server) {
	req, err := http.NewRequest("GET", ts.URL+"/api/checkauth", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("apiKey", apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Error("Failed to validate api key: ", apiKey)
	}
}

func invalidAPIKeyTestComponentInfo(t *testing.T, ts *httptest.Server) {
	req, err := http.NewRequest("GET", ts.URL+"/api/checkauth", nil)
	invalidKey := "invalidAPIKey"
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("apiKey", invalidKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Error("Invalid API key was accepted: ", invalidKey)
	}
}
