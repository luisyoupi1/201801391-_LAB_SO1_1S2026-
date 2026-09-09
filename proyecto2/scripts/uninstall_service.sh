#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-201801391}"
TARGET_ROOT="/opt/so1-proyecto2-${CARNET}"
systemctl disable --now so1-telemetryd.service 2>/dev/null || true
if [ -f "$TARGET_ROOT/docker-compose.yml" ]; then
  if docker compose version >/dev/null 2>&1; then
    docker compose --env-file "$TARGET_ROOT/.env" -f "$TARGET_ROOT/docker-compose.yml" down || true
  elif command -v docker-compose >/dev/null 2>&1; then
    docker-compose --env-file "$TARGET_ROOT/.env" -f "$TARGET_ROOT/docker-compose.yml" down || true
  fi
fi
bash "${TARGET_ROOT}/scripts/remove_cron.sh" 2>/dev/null || true
bash "${TARGET_ROOT}/scripts/unload_module.sh" 2>/dev/null || true
rm -f /etc/systemd/system/so1-telemetryd.service
rm -f "/etc/so1-proyecto2-${CARNET}.env" "/usr/local/bin/so1-telemetryd-${CARNET}"
rm -rf -- "$TARGET_ROOT"
systemctl daemon-reload
echo "Servicio retirado. Los volúmenes de Docker se conservaron."

