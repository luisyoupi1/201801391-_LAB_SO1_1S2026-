CARNET ?= 201801391
TAG ?= 1.0.0
GOFLAGS ?= -buildvcs=false
export GOFLAGS

.PHONY: test build fmt vet compose-up compose-down

test:
	go test ./...

build:
	go build -o bin/api1 ./api1
	go build -o bin/api2 ./api2
	go build -o bin/api3 ./api3

fmt:
	gofmt -w api1 api2 api3 internal

vet:
	go vet ./...

compose-up:
	CARNET=$(CARNET) TAG=$(TAG) docker compose up --build -d

compose-down:
	docker compose down
