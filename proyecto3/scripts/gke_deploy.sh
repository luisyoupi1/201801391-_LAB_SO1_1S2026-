#!/usr/bin/env bash
set -euo pipefail

if [[ "${CONFIRM_GCP_DEPLOY:-no}" != "yes" ]]; then
  echo "No se desplegó. Para autorizar cargos use CONFIRM_GCP_DEPLOY=yes." >&2
  exit 2
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
: "${IMAGE_REGISTRY:?Defina IMAGE_REGISTRY, por ejemplo zot.midominio.com}"
render() {
  kubectl kustomize "$1" | sed "s#zot\.example\.com:5000#$IMAGE_REGISTRY#g"
}
render "$ROOT/operator/config" | kubectl apply -f -
kubectl rollout status deployment/mum-operator -n mum-system --timeout=180s
render "$ROOT/k8s/gke" | kubectl apply -f -
kubectl wait --for=condition=Available deployment --all -n so1-proyecto3 --timeout=300s
kubectl get gateway,httproute,pods,svc,hpa -n so1-proyecto3
