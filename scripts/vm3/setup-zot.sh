#!/usr/bin/env bash
set -euo pipefail

PROJECT_DIR="${PROJECT_DIR:-/proyecto}"
REGISTRY="${REGISTRY:-192.168.56.13:5000}"
ZOT_IMAGE="${ZOT_IMAGE:-ghcr.io/project-zot/zot-linux-amd64:latest}"

export DEBIAN_FRONTEND=noninteractive
sudo apt-get update
sudo apt-get install -y ca-certificates curl docker.io jq

sudo install -d -m 0755 /etc/docker
if [[ -s /etc/docker/daemon.json ]]; then
  temp_daemon="$(mktemp)"
  sudo jq --arg registry "$REGISTRY" \
    '. + {"insecure-registries": (((.["insecure-registries"] // []) + [$registry]) | unique)}' \
    /etc/docker/daemon.json > "$temp_daemon"
  sudo install -m 0644 "$temp_daemon" /etc/docker/daemon.json
  rm -f -- "$temp_daemon"
else
  printf '{"insecure-registries":["%s"]}\n' "$REGISTRY" | sudo tee /etc/docker/daemon.json >/dev/null
fi

sudo systemctl enable --now docker
sudo systemctl restart docker

sudo install -d -m 0755 /opt/zot
sudo install -m 0644 "${PROJECT_DIR}/infra/zot/config.json" /opt/zot/config.json
sudo docker rm -f zot >/dev/null 2>&1 || true
sudo docker run -d --name zot --restart always \
  -p 5000:5000 \
  -v /opt/zot/config.json:/etc/zot/config.json:ro \
  -v zot-data:/var/lib/registry \
  "$ZOT_IMAGE" serve /etc/zot/config.json

for _ in $(seq 1 30); do
  if curl --fail --silent "http://127.0.0.1:5000/v2/" >/dev/null; then
    echo "Zot está disponible en http://${REGISTRY}/v2/."
    exit 0
  fi
  sleep 1
done

echo "Zot no respondió después de 30 segundos." >&2
sudo docker logs zot >&2
exit 1
