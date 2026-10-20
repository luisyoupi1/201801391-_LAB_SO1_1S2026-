#!/usr/bin/env bash
set -euo pipefail

wait_url() {
  local name="$1" url="$2"
  for _ in $(seq 1 90); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      printf '[OK] %s: %s\n' "$name" "$url"
      return 0
    fi
    sleep 2
  done
  printf '[FALLO] %s no respondió: %s\n' "$name" "$url" >&2
  return 1
}

wait_url "API Rust" "http://localhost:${RUST_API_PORT:-8080}/health"
wait_url "Prometheus" "http://localhost:${PROMETHEUS_PORT:-9090}/-/ready"
wait_url "Grafana" "http://localhost:${GRAFANA_PORT:-3000}/api/health"
echo "Proyecto 3 listo en local."
