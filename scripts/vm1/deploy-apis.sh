#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/proyecto}"
CARNET="${CARNET:-201801391}"
TAG="${TAG:-1.0.0}"
REGISTRY="${REGISTRY:-192.168.56.13:5000}"
VM1_IP="${VM1_IP:-192.168.56.11}"
VM2_IP="${VM2_IP:-192.168.56.12}"

for api in api1 api2; do
  image="${REGISTRY}/${api}-${CARNET}:${TAG}"
  sudo nerdctl build -f "${PROJECT_DIR}/${api}/Dockerfile" -t "$image" "$PROJECT_DIR"
  sudo nerdctl --insecure-registry push "$image"
  sudo nerdctl image rm "$image" >/dev/null
  sudo nerdctl --insecure-registry pull "$image"
done

sudo nerdctl rm -f api1 >/dev/null 2>&1 || true
sudo nerdctl rm -f api2 >/dev/null 2>&1 || true

sudo nerdctl run -d --name api1 --restart always --net host \
  -e CARNET="$CARNET" -e VM_NAME=VM1 -e PORT=8081 \
  -e API2_URL="http://${VM1_IP}:8082" -e API3_URL="http://${VM2_IP}:8083" \
  "${REGISTRY}/api1-${CARNET}:${TAG}"

sudo nerdctl run -d --name api2 --restart always --net host \
  -e CARNET="$CARNET" -e VM_NAME=VM1 -e PORT=8082 \
  -e API1_URL="http://${VM1_IP}:8081" -e API3_URL="http://${VM2_IP}:8083" \
  "${REGISTRY}/api2-${CARNET}:${TAG}"

sleep 2
curl --fail --silent --show-error "http://127.0.0.1:8081/health" | jq .
curl --fail --silent --show-error "http://127.0.0.1:8082/health" | jq .

echo "API1 y API2 desplegadas con Containerd en VM1."
