package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	pb "github.com/luisyoupi1/so1-proyecto3/services-go/gen/wartweets/v1"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func country(value string) (pb.Countries, error) {
	if number, ok := pb.Countries_value[value]; ok {
		return pb.Countries(number), nil
	}
	return pb.Countries_COUNTRIES_UNKNOWN, fmt.Errorf("unknown country %s", value)
}

func main() {
	conn, err := grpc.NewClient(config.Env("GRPC_SERVER_ADDR", "grpc-server:50051"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWarReportServiceClient(conn)
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "ok") })
	mux.HandleFunc("/forward", func(w http.ResponseWriter, r *http.Request) {
		var report model.WarReport
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&report); err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		countryValue, err := country(report.Country)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		result, err := client.SendReport(ctx, &pb.WarReportRequest{Id: report.ID, Country: countryValue, WarplanesInAir: report.WarplanesInAir, WarshipsInWater: report.WarshipsInWater, Timestamp: report.Timestamp.Format(time.RFC3339Nano)})
		if err != nil {
			http.Error(w, err.Error(), 502)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(result)
	})
	addr := config.Env("HTTP_ADDR", ":8082")
	log.Printf("grpc-client listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
