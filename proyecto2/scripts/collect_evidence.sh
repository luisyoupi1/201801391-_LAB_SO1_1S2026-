#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-201801391}"
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
OUT="$PROJECT_ROOT/evidencias"
mkdir -p "$OUT"

uname -a > "$OUT/01-kernel.txt"
lsmod | grep '^continfo' > "$OUT/02-modulo-cargado.txt"
cat "/proc/continfo_pr2_so1_${CARNET}" | jq . > "$OUT/03-proc-snapshot.json"
docker ps --filter "label=so1.project=${CARNET}" --no-trunc > "$OUT/04-contenedores.txt"
curl -fsS http://127.0.0.1:9105/metrics > "$OUT/05-prometheus-metrics.txt"
docker exec "so1-valkey-${CARNET}" valkey-cli --scan --pattern "so1:${CARNET}:*" > "$OUT/06-valkey-keys.txt"
journalctl -u so1-telemetryd --no-pager -n 200 > "$OUT/07-daemon.log"
curl -fsS http://127.0.0.1:3000/api/health > "$OUT/08-grafana-health.json"

echo "Evidencias guardadas en $OUT"

