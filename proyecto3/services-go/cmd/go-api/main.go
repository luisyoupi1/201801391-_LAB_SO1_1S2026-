package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
)

func main() {
	client := &http.Client{Timeout: 10 * time.Second}
	forwardURL := config.Env("GRPC_CLIENT_URL", "http://grpc-client:8082/forward")
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	mux.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var report model.WarReport
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&report); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := report.Validate(); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		payload, _ := json.Marshal(report)
		request, _ := http.NewRequestWithContext(r.Context(), http.MethodPost, forwardURL, bytes.NewReader(payload))
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		defer response.Body.Close()
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
	})
	addr := config.Env("HTTP_ADDR", ":8081")
	log.Printf("go-api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
