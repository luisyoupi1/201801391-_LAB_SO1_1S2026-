#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${REGISTRY:-192.168.56.13:5000}"

export DEBIAN_FRONTEND=noninteractive
sudo apt-get update
sudo apt-get install -y ca-certificates curl podman

sudo install -d -m 0755 /etc/containers/registries.conf.d
sudo tee /etc/containers/registries.conf.d/so1-registry.conf >/dev/null <<EOF
[[registry]]
location = "${REGISTRY}"
insecure = true
EOF

sudo podman info --format '{{.Host.OCIRuntime.Name}}'
echo "Podman quedó listo para el registro ${REGISTRY}."
