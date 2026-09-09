FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY internal ./internal
COPY api1 ./api1
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./api1

FROM scratch
COPY --from=build /out/api /api
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/api"]
