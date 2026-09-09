#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-201801391}"
PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
CRON_FILE="/etc/cron.d/so1-proyecto2-${CARNET}"
LOG_FILE="/var/log/so1-proyecto2-${CARNET}-cron.log"

touch "$LOG_FILE"
chmod 0644 "$LOG_FILE"
chmod 0755 "$PROJECT_ROOT/scripts/generate_containers.sh"
cat > "$CRON_FILE" <<EOF
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
CARNET=${CARNET}
PROJECT_ROOT=${PROJECT_ROOT}
*/2 * * * * root /bin/bash ${PROJECT_ROOT}/scripts/generate_containers.sh >> ${LOG_FILE} 2>&1
EOF
chmod 0644 "$CRON_FILE"
systemctl reload cron 2>/dev/null || systemctl restart cron

