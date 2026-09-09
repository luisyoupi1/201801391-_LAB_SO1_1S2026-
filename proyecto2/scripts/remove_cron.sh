#!/usr/bin/env bash
set -euo pipefail
CARNET="${CARNET:-201801391}"
rm -f "/etc/cron.d/so1-proyecto2-${CARNET}"
systemctl reload cron 2>/dev/null || true

