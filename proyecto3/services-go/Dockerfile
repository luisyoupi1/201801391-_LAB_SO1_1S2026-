FROM golang:1.24-bookworm AS tooling
RUN apt-get update && apt-get install -y --no-install-recommends protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.6 \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN protoc -I proto \
    --go_out=. --go_opt=module=github.com/luisyoupi1/so1-proyecto3/services-go \
    --go-grpc_out=. --go-grpc_opt=module=github.com/luisyoupi1/so1-proyecto3/services-go \
    proto/warreport.proto

FROM tooling AS test
RUN go test -race ./... && go vet ./...

FROM tooling AS build
ARG SERVICE
RUN test -n "$SERVICE" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/service "./cmd/${SERVICE}"

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/service /service
USER nonroot:nonroot
ENTRYPOINT ["/service"]
