#!/usr/bin/env bash
set -euo pipefail

for _ in $(seq 1 60); do
  if curl -kfsS "https://localhost:${ZOT_PORT:-5000}/v2/" >/dev/null 2>&1; then
    echo "[OK] Zot HTTPS: https://localhost:${ZOT_PORT:-5000}/v2/"
    exit 0
  fi
  sleep 2
done
echo "[FALLO] Zot no respondió" >&2
exit 1
