#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

curl -fsS http://localhost:8080/health | python3 -m json.tool >/dev/null
before="$(docker compose --env-file .env exec -T valkey valkey-cli GET mum:reports:total | tr -d '\r')"
before="${before:-0}"

payload='{"country":"USA","warplanes_in_air":17,"warships_in_water":9,"timestamp":"2026-09-28T12:00:00Z"}'
response="$(curl -fsS -H 'Content-Type: application/json' -d "$payload" http://localhost:8080/grpc-201801391)"
python3 -c 'import json,sys; assert json.load(sys.stdin)["status"] == "accepted"' <<<"$response"

for _ in $(seq 1 30); do
  after="$(docker compose --env-file .env exec -T valkey valkey-cli GET mum:reports:total | tr -d '\r')"
  after="${after:-0}"
  if (( after > before )); then break; fi
  sleep 1
done
(( after > before )) || { echo "[FALLO] El reporte no llegó a Valkey" >&2; exit 1; }

curl -fsS http://localhost:9106/metrics | grep -q 'mum_reports_total'
curl -fsS 'http://localhost:9090/api/v1/query?query=mum_reports_total' | grep -q '"status":"success"'
curl -fsS http://localhost:3000/api/health | grep -q '"database"'
echo "[OK] Flujo completo: Rust -> Go -> gRPC -> RabbitMQ -> Consumer -> Valkey -> Prometheus -> Grafana"
