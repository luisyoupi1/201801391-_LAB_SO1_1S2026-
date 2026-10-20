package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/storage"
)

type cloudEvent struct {
	Data json.RawMessage `json:"data"`
}

func main() {
	store := storage.New(config.Env("VALKEY_ADDR", "valkey:6379"))
	defer store.Close()
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	mux.HandleFunc("/dapr/subscribe", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `[{"pubsubname":"rabbitmq-pubsub","topic":"war-reports","route":"/dapr/war-reports"}]`)
	})
	mux.HandleFunc("/dapr/war-reports", func(w http.ResponseWriter, r *http.Request) {
		var event cloudEvent
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&event); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		var report model.WarReport
		if err := json.Unmarshal(event.Data, &report); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := store.Save(r.Context(), report); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"SUCCESS"}`)
	})
	addr := config.Env("HTTP_ADDR", ":8084")
	log.Printf("dapr-consumer listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
