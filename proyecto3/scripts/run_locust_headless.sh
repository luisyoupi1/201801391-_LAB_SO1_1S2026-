#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
mkdir -p runtime/locust
docker compose --env-file .env -f docker-compose.yml --profile load run --rm \
  locust -f /mnt/locust/locustfile.py --host http://rust-api:8080 \
  --headless -u "${USERS:-10}" -r "${SPAWN_RATE:-2}" -t "${DURATION:-30s}" \
  --csv /reports/results --html /reports/report.html
echo "Reporte: $ROOT/runtime/locust/report.html"
