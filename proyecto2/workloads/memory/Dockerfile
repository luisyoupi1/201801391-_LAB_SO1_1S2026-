FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/memory-workload .

FROM alpine:3.21
RUN addgroup -S workload && adduser -S -G workload workload
COPY --from=build /out/memory-workload /usr/local/bin/memory-workload
USER workload
ENTRYPOINT ["/usr/local/bin/memory-workload"]

