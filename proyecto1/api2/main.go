package main

import (
	"log"

	"sopes1_proyecto1_201801391/internal/service"
)

func main() {
	carnet := service.Env("CARNET", "201801391")
	config := service.Config{
		Name:    "API2",
		VM:      service.Env("VM_NAME", "VM1"),
		Carnet:  carnet,
		Address: ":" + service.Env("PORT", "8082"),
		Targets: []service.Target{
			{Name: "API1", VM: "VM1", BaseURL: service.Env("API1_URL", "http://api1:8081")},
			{Name: "API3", VM: "VM2", BaseURL: service.Env("API3_URL", "http://api3:8083")},
		},
	}
	if err := service.Serve(config); err != nil {
		log.Fatal(err)
	}
}
