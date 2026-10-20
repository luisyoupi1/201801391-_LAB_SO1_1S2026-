package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"github.com/redis/go-redis/v9"
)

type exporter struct{ client *redis.Client }

func parse(value string) float64 { result, _ := strconv.ParseFloat(value, 64); return result }

func (e exporter) metrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := e.client.Ping(ctx).Err(); err != nil {
		http.Error(w, err.Error(), 503)
		return
	}
	global, _ := e.client.HGetAll(ctx, "mum:global").Result()
	counts, _ := e.client.HGetAll(ctx, "mum:country:counts").Result()
	planesSum, _ := e.client.HGetAll(ctx, "mum:country:planes_sum").Result()
	shipsSum, _ := e.client.HGetAll(ctx, "mum:country:ships_sum").Result()
	planesLatest, _ := e.client.HGetAll(ctx, "mum:country:planes_latest").Result()
	shipsLatest, _ := e.client.HGetAll(ctx, "mum:country:ships_latest").Result()
	total, _ := e.client.Get(ctx, "mum:reports:total").Result()
	planesMode, _ := e.client.ZRevRangeWithScores(ctx, "mum:freq:planes", 0, 0).Result()
	shipsMode, _ := e.client.ZRevRangeWithScores(ctx, "mum:freq:ships", 0, 0).Result()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintln(w, "# TYPE mum_reports_total counter")
	fmt.Fprintf(w, "mum_reports_total %s\n", zero(total))
	for _, key := range []string{"planes_max", "planes_min", "ships_max", "ships_min"} {
		fmt.Fprintf(w, "mum_%s %s\n", key, zero(global[key]))
	}
	if len(planesMode) > 0 {
		fmt.Fprintf(w, "mum_planes_mode %v\n", parse(fmt.Sprint(planesMode[0].Member)))
	} else {
		fmt.Fprintln(w, "mum_planes_mode 0")
	}
	if len(shipsMode) > 0 {
		fmt.Fprintf(w, "mum_ships_mode %v\n", parse(fmt.Sprint(shipsMode[0].Member)))
	} else {
		fmt.Fprintln(w, "mum_ships_mode 0")
	}
	for _, country := range model.Countries {
		fmt.Fprintf(w, "mum_country_reports_total{country=%q} %s\n", country, zero(counts[country]))
		fmt.Fprintf(w, "mum_country_warplanes_total{country=%q} %s\n", country, zero(planesSum[country]))
		fmt.Fprintf(w, "mum_country_warships_total{country=%q} %s\n", country, zero(shipsSum[country]))
		fmt.Fprintf(w, "mum_country_warplanes_current{country=%q} %s\n", country, zero(planesLatest[country]))
		fmt.Fprintf(w, "mum_country_warships_current{country=%q} %s\n", country, zero(shipsLatest[country]))
	}
}

func zero(value string) string {
	if value == "" || value == "redis: nil" {
		return "0"
	}
	return value
}

func main() {
	client := redis.NewClient(&redis.Options{Addr: config.Env("VALKEY_ADDR", "valkey:6379")})
	defer client.Close()
	e := exporter{client: client}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := client.Ping(r.Context()).Err(); err != nil {
			http.Error(w, err.Error(), 503)
			return
		}
		_, _ = fmt.Fprint(w, "ok")
	})
	mux.HandleFunc("/metrics", e.metrics)
	addr := config.Env("HTTP_ADDR", ":9106")
	log.Printf("metrics-exporter listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
