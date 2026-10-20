package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	amqp "github.com/rabbitmq/amqp091-go"
)

type publisher struct {
	mu         sync.Mutex
	url, queue string
	conn       *amqp.Connection
	channel    *amqp.Channel
}

func (p *publisher) connect() error {
	if p.conn != nil && !p.conn.IsClosed() && p.channel != nil {
		return nil
	}
	conn, err := amqp.Dial(p.url)
	if err != nil {
		return err
	}
	channel, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return err
	}
	if _, err = channel.QueueDeclare(p.queue, true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = conn.Close()
		return err
	}
	p.conn, p.channel = conn, channel
	return nil
}

func (p *publisher) publish(ctx context.Context, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if err := p.connect(); err != nil {
		return err
	}
	err := p.channel.PublishWithContext(ctx, "", p.queue, false, false, amqp.Publishing{DeliveryMode: amqp.Persistent, ContentType: "application/json", Timestamp: time.Now(), Body: payload})
	if err != nil {
		_ = p.conn.Close()
		p.conn, p.channel = nil, nil
	}
	return err
}

func main() {
	p := &publisher{url: config.Env("AMQP_URL", "amqp://guest:guest@rabbitmq:5672/"), queue: config.Env("RABBITMQ_QUEUE", "war-reports")}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	mux.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
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
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := p.publish(ctx, payload); err != nil {
			http.Error(w, err.Error(), 503)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = io.WriteString(w, `{"status":"queued"}`)
	})
	addr := config.Env("HTTP_ADDR", ":8083")
	log.Printf("rabbit-writer listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
