package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const apiKey = "secretKey"

type CheckAuthResponse struct {
	Valid bool
	Error string
}

func main() {
	// Routes
	http.HandleFunc("/api/checkauth", checkAuth)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Something went wrong: ", err)
		return
	}
}

// The Test URL you can use to test your API key with the provider using the "Test" button
func checkAuth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("checkAuth")
	resp := CheckAuthResponse{true, ""}
	key := r.Header.Get("apiKey")
	if key != apiKey {
		fmt.Println("not working", key)
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
