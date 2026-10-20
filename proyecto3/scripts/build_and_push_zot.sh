#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="${ZOT_REGISTRY:-127.0.0.1:5000}"
TAG="${IMAGE_TAG:-v1}"

cd "$ROOT"
docker compose --env-file .env -f docker-compose.yml build
docker build -t localhost:5000/so1/dapr-consumer:local --build-arg SERVICE=dapr-consumer services-go
docker build -t localhost:5000/so1/mum-operator:local operator

publish() {
  local service="$1"
  local source="localhost:5000/so1/$service:local"
  local target="$REGISTRY/so1/$service:$TAG"
  docker tag "$source" "$target"
  if [[ "$REGISTRY" == "127.0.0.1:5000" ]]; then
    docker run --rm --network host -v /var/run/docker.sock:/var/run/docker.sock \
      quay.io/skopeo/stable:v1.18.0 copy --dest-tls-verify=false \
      "docker-daemon:$target" "docker://$target"
  else
    docker push "$target"
  fi
}

for service in rust-api go-api grpc-client grpc-server rabbit-writer consumer metrics-exporter dapr-consumer; do
  publish "$service"
done
publish mum-operator
echo "[OK] Imágenes publicadas en Zot local: $REGISTRY/so1/*:$TAG"
