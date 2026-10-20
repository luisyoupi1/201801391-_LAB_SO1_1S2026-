package main

import (
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/storage"
	amqp "github.com/rabbitmq/amqp091-go"
)

func consume(ctx context.Context, url, queue string, store *storage.Store) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		return err
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()
	if err = channel.Qos(32, 0, false); err != nil {
		return err
	}
	if _, err = channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		return err
	}
	deliveries, err := channel.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return amqp.ErrClosed
			}
			var report model.WarReport
			if err := json.Unmarshal(delivery.Body, &report); err != nil || report.Validate() != nil {
				_ = delivery.Nack(false, false)
				continue
			}
			if err := store.Save(ctx, report); err != nil {
				log.Printf("store report %s: %v", report.ID, err)
				_ = delivery.Nack(false, true)
				continue
			}
			_ = delivery.Ack(false)
			log.Printf("stored report id=%s country=%s", report.ID, report.Country)
		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	store := storage.New(config.Env("VALKEY_ADDR", "valkey:6379"))
	defer store.Close()
	url, queue := config.Env("AMQP_URL", "amqp://guest:guest@rabbitmq:5672/"), config.Env("RABBITMQ_QUEUE", "war-reports")
	for attempt := 0; ctx.Err() == nil; attempt++ {
		if err := consume(ctx, url, queue, store); err != nil {
			log.Printf("consumer reconnect: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(config.Retry(attempt)):
			}
		} else {
			return
		}
	}
}
