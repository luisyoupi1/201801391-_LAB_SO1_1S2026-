#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
: "${IMAGE_REGISTRY:?Defina IMAGE_REGISTRY, por ejemplo zot.midominio.com}"
helm repo add dapr https://dapr.github.io/helm-charts/
helm repo update
helm upgrade --install dapr dapr/dapr --namespace dapr-system --create-namespace --wait
kubectl kustomize "$ROOT/k8s/dapr" | sed "s#zot\.example\.com:5000#$IMAGE_REGISTRY#g" | kubectl apply -f -
