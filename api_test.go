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

	t.Run("Valid Api Key", func(t *testing.T) {
		validAPIKeyTest(t, ts)
	})
	t.Run("Invalid Api Key", func(t *testing.T) {
		invalidAPIKeyTest(t, ts)
	})

	//err = json.Unmarshal(body, &rr)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Error(rr)

	//tests := []struct {
	//	name string
	//	r    *http.Request
	//}{
	//	{name: "1: testing get", r: newreq("GET", ts.URL+"/checkauth", nil)},
	//}
	//for _, test := range tests {
	//	t.Run(test.name, func(t *testing.T) {
	//		resp, err := http.DefaultClient.Do(tt.r)
	//
	//		defer resp.Body.Close()
	//
	//		if err != nil {
	//			t.Fatal(err)
	//		}
	//		// check for expected response here.
	//	})
	//}
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
		t.Error("Failed to validate api key:")
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
