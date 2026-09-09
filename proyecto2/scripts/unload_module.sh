#!/usr/bin/env bash
set -euo pipefail
if lsmod | awk '{print $1}' | grep -qx continfo; then
  rmmod continfo
fi

