#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/proyecto}"
CARNET="${CARNET:-201801391}"
TAG="${TAG:-1.0.0}"
REGISTRY="${REGISTRY:-192.168.56.13:5000}"
VM1_IP="${VM1_IP:-192.168.56.11}"

image="${REGISTRY}/api3-${CARNET}:${TAG}"
podman build -f "${PROJECT_DIR}/api3/Dockerfile" -t "$image" "$PROJECT_DIR"
podman push --tls-verify=false "$image"
podman image rm "$image"
podman pull --tls-verify=false "$image"

podman rm -f api3 >/dev/null 2>&1 || true
podman run -d --name api3 --restart always --network host \
  -e CARNET="$CARNET" -e VM_NAME=VM2 -e PORT=8083 \
  -e API1_URL="http://${VM1_IP}:8081" -e API2_URL="http://${VM1_IP}:8082" \
  "$image"

sleep 2
curl --fail --silent --show-error "http://127.0.0.1:8083/health" | jq .
echo "API3 desplegada con Podman en VM2."
