package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "github.com/luisyoupi1/so1-proyecto3/services-go/gen/wartweets/v1"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/config"
	"github.com/luisyoupi1/so1-proyecto3/services-go/internal/model"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedWarReportServiceServer
	client    *http.Client
	writerURL string
}

func (s *server) SendReport(ctx context.Context, request *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	timestamp, err := time.Parse(time.RFC3339Nano, request.GetTimestamp())
	if err != nil {
		return nil, fmt.Errorf("timestamp: %w", err)
	}
	report := model.WarReport{ID: request.GetId(), Country: request.GetCountry().String(), WarplanesInAir: request.GetWarplanesInAir(), WarshipsInWater: request.GetWarshipsInWater(), Timestamp: timestamp}
	if err := report.Validate(); err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(report)
	httpRequest, _ := http.NewRequestWithContext(ctx, http.MethodPost, s.writerURL, bytes.NewReader(payload))
	httpRequest.Header.Set("Content-Type", "application/json")
	response, err := s.client.Do(httpRequest)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode/100 != 2 {
		return nil, fmt.Errorf("writer returned %s", response.Status)
	}
	return &pb.WarReportResponse{Status: "published", Id: report.ID}, nil
}

func main() {
	listener, err := net.Listen("tcp", config.Env("GRPC_ADDR", ":50051"))
	if err != nil {
		log.Fatal(err)
	}
	grpcServer := grpc.NewServer()
	pb.RegisterWarReportServiceServer(grpcServer, &server{client: &http.Client{Timeout: 8 * time.Second}, writerURL: config.Env("WRITER_URL", "http://rabbit-writer:8083/publish")})
	log.Printf("grpc-server listening on %s", listener.Addr())
	log.Fatal(grpcServer.Serve(listener))
}
