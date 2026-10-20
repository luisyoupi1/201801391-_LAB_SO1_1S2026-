#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="${ZOT_REGISTRY:-127.0.0.1:5000}"
TAG="${OCI_TAG:-v1}"
cert_mount=()
tls_args=()
if [[ "$REGISTRY" == "127.0.0.1:5000" ]]; then
  cert_mount=(-v "$ROOT/certs:/certs:ro")
  tls_args=(--ca-file /certs/zot.crt)
fi

docker run --rm --network host \
  -v "$ROOT/artifacts:/artifacts:ro" \
  "${cert_mount[@]}" \
  -w /artifacts \
  ghcr.io/oras-project/oras:v1.2.3 push "${tls_args[@]}" \
  "$REGISTRY/so1/countries:$TAG" \
  --artifact-type application/vnd.so1.countries.v1+json \
  countries.json:application/json

echo "[OK] Artefacto OCI disponible en $REGISTRY/so1/countries:$TAG"
ZOT_REGISTRY="$REGISTRY" OCI_TAG="$TAG" "$ROOT/scripts/pull_oci_artifact.sh"
