#!/usr/bin/env bash
set -euo pipefail

REGISTRY="${REGISTRY:-192.168.56.13:5000}"
NERDCTL_VERSION="${NERDCTL_VERSION:-2.0.3}"

export DEBIAN_FRONTEND=noninteractive
sudo apt-get update
sudo apt-get install -y ca-certificates containerd curl tar

if ! command -v nerdctl >/dev/null 2>&1 || ! command -v buildkitd >/dev/null 2>&1; then
  case "$(uname -m)" in
    x86_64) nerdctl_arch="amd64" ;;
    aarch64|arm64) nerdctl_arch="arm64" ;;
    *) echo "Arquitectura no soportada para la instalación automática de nerdctl." >&2; exit 1 ;;
  esac
  temp_dir="$(mktemp -d)"
  trap 'rm -rf -- "$temp_dir"' EXIT
  curl --fail --location --show-error \
    "https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-full-${NERDCTL_VERSION}-linux-${nerdctl_arch}.tar.gz" \
    --output "${temp_dir}/nerdctl-full.tar.gz"
  tar -xzf "${temp_dir}/nerdctl-full.tar.gz" -C "$temp_dir" \
    bin/nerdctl bin/buildctl bin/buildkitd
  sudo install -m 0755 "${temp_dir}/bin/nerdctl" /usr/local/bin/nerdctl
  sudo install -m 0755 "${temp_dir}/bin/buildctl" /usr/local/bin/buildctl
  sudo install -m 0755 "${temp_dir}/bin/buildkitd" /usr/local/bin/buildkitd
fi

sudo install -d -m 0755 /etc/containerd
containerd config default | sudo tee /etc/containerd/config.toml >/dev/null
sudo sed -i 's|config_path = ""|config_path = "/etc/containerd/certs.d"|' /etc/containerd/config.toml

sudo install -d -m 0755 "/etc/containerd/certs.d/${REGISTRY}"
sudo tee "/etc/containerd/certs.d/${REGISTRY}/hosts.toml" >/dev/null <<EOF
server = "http://${REGISTRY}"

[host."http://${REGISTRY}"]
  capabilities = ["pull", "resolve", "push"]
  skip_verify = true
EOF

sudo systemctl enable --now containerd
sudo systemctl restart containerd

sudo tee /etc/systemd/system/buildkit.service >/dev/null <<'EOF'
[Unit]
Description=BuildKit
Requires=containerd.service
After=containerd.service

[Service]
ExecStart=/usr/local/bin/buildkitd --oci-worker=false --containerd-worker=true
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now buildkit
sudo nerdctl version
sudo buildctl debug workers

echo "Containerd, nerdctl y BuildKit quedaron listos para el registro ${REGISTRY}."
