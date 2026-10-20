#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
helm repo add falcosecurity https://falcosecurity.github.io/charts
helm repo update
helm upgrade --install falco falcosecurity/falco --namespace falco --create-namespace \
  -f "$ROOT/k8s/falco/values.yaml" --wait
echo "Evidencia: kubectl logs -n falco -l app.kubernetes.io/name=falco -c falco | grep 'MUM sensitive'"
