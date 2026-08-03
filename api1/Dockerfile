FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY api1 ./api1
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./api1

FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app
COPY --from=build /out/api /usr/local/bin/api
USER app
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/api"]
