#!/usr/bin/env bash
set -euo pipefail

if [ "${EUID}" -ne 0 ]; then
  echo "Ejecute: sudo bash scripts/install_dependencies.sh" >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y \
  build-essential clang llvm libbpf-dev linux-headers-"$(uname -r)" \
  linux-tools-common linux-tools-generic linux-tools-"$(uname -r)" \
  golang-go docker.io cron curl jq python3 rsync ca-certificates

if ! docker compose version >/dev/null 2>&1; then
  if ! apt-get install -y docker-compose-plugin; then
    apt-get install -y docker-compose
  fi
fi

systemctl enable --now docker cron
if [ -n "${SUDO_USER:-}" ] && [ "${SUDO_USER}" != "root" ]; then
  usermod -aG docker "$SUDO_USER"
  echo "El usuario $SUDO_USER fue agregado al grupo docker; cierre sesión y vuelva a entrar."
fi

if [ ! -r /sys/kernel/btf/vmlinux ]; then
  echo "ADVERTENCIA: /sys/kernel/btf/vmlinux no existe; la compilación eBPF requiere BTF." >&2
fi

echo "Dependencias instaladas. Versiones:"
go version
clang --version | head -n 1
docker --version

