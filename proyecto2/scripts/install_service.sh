#!/usr/bin/env bash
set -euo pipefail

if [ "${EUID}" -ne 0 ]; then
  echo "Este instalador requiere root" >&2
  exit 1
fi

CARNET="${CARNET:-201801391}"
SOURCE_ROOT="${PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
TARGET_ROOT="/opt/so1-proyecto2-${CARNET}"
ENV_FILE="/etc/so1-proyecto2-${CARNET}.env"
SERVICE_FILE="/etc/systemd/system/so1-telemetryd.service"

test -x "$SOURCE_ROOT/bin/telemetryd"
test -f "$SOURCE_ROOT/kernel/continfo.ko"
test -f "$SOURCE_ROOT/daemon/internal/ebpf/kill_monitor.bpf.o"

install -d -m 0755 "$TARGET_ROOT"
rsync -a --exclude .git --exclude evidencias/ "$SOURCE_ROOT/" "$TARGET_ROOT/"
chmod 0755 "$TARGET_ROOT/scripts/"*.sh
if [ ! -f "$TARGET_ROOT/.env" ]; then
  cp "$TARGET_ROOT/.env.example" "$TARGET_ROOT/.env"
fi
install -m 0755 "$SOURCE_ROOT/bin/telemetryd" "/usr/local/bin/so1-telemetryd-${CARNET}"

cat > "$ENV_FILE" <<EOF
CARNET=${CARNET}
PROC_PATH=/proc/continfo_pr2_so1_${CARNET}
LOOP_INTERVAL=30s
METRICS_ADDR=:9105
VALKEY_ADDR=127.0.0.1:6379
VALKEY_PASSWORD=
COMPOSE_FILE=${TARGET_ROOT}/docker-compose.yml
PROJECT_ROOT=${TARGET_ROOT}
BPF_OBJECT=${TARGET_ROOT}/daemon/internal/ebpf/kill_monitor.bpf.o
MIN_LOW_CONTAINERS=3
MIN_HIGH_CONTAINERS=2
DELETE_CONFIRM_TIMEOUT=15s
EBPF_REQUIRED=true
EOF
chmod 0600 "$ENV_FILE"

sed "s/201801391/${CARNET}/g" "$SOURCE_ROOT/systemd/so1-telemetryd.service" > "$SERVICE_FILE"
systemctl daemon-reload
systemctl enable --now so1-telemetryd.service
echo "Servicio instalado. Consulte: systemctl status so1-telemetryd"
