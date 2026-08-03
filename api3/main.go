package main

import (
	"log"

	"sopes1_proyecto1_201801391/internal/service"
)

func main() {
	carnet := service.Env("CARNET", "201801391")
	config := service.Config{
		Name:    "API3",
		VM:      service.Env("VM_NAME", "VM2"),
		Carnet:  carnet,
		Address: ":" + service.Env("PORT", "8083"),
		Targets: []service.Target{
			{Name: "API1", VM: "VM1", BaseURL: service.Env("API1_URL", "http://api1:8081")},
			{Name: "API2", VM: "VM1", BaseURL: service.Env("API2_URL", "http://api2:8082")},
		},
	}
	if err := service.Serve(config); err != nil {
		log.Fatal(err)
	}
}
