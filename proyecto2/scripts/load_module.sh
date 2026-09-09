#!/usr/bin/env bash
set -euo pipefail

PROJECT_ROOT="${PROJECT_ROOT:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
MODULE="$PROJECT_ROOT/kernel/continfo.ko"

if [ ! -f "$MODULE" ]; then
  echo "No existe $MODULE; ejecute make kernel" >&2
  exit 1
fi
if lsmod | awk '{print $1}' | grep -qx continfo; then
  rmmod continfo
fi
insmod "$MODULE"
test -r "/proc/continfo_pr2_so1_${CARNET:-201801391}"
dmesg | tail -n 5

