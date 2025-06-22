package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

var started = time.Now()

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)

	fmt.Println("Listening on :8000")
	err := http.ListenAndServe("0.0.0.0:8000", mux)
	fmt.Println("Shutting down")
	if err != nil {
		fmt.Println(err.Error())
	}
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	if time.Since(started).Seconds() < 10 {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Service unavailable"))

		return
	}

	marshal, err := json.Marshal(HealthResponse{
		Status: "OK",
		Host:   os.Getenv("HOSTNAME"),
	})
	if err != nil {
		fmt.Println(err.Error())

		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		fmt.Println(err.Error())
	}
}

type HealthResponse struct {
	Status string `json:"status"`
	Host   string `json:"host"`
}
