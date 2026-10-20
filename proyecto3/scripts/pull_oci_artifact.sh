#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="${ZOT_REGISTRY:-127.0.0.1:5000}"
TAG="${OCI_TAG:-v1}"
OUTPUT="$ROOT/runtime/oci"
cert_mount=()
tls_args=()

if [[ "$REGISTRY" == "127.0.0.1:5000" ]]; then
  cert_mount=(-v "$ROOT/certs:/certs:ro")
  tls_args=(--ca-file /certs/zot.crt)
fi

mkdir -p "$OUTPUT"
find "$OUTPUT" -mindepth 1 -maxdepth 1 -type f -delete

docker run --rm --network host \
  -v "$OUTPUT:/output" \
  "${cert_mount[@]}" \
  ghcr.io/oras-project/oras:v1.2.3 pull "${tls_args[@]}" \
  "$REGISTRY/so1/countries:$TAG" --output /output

python3 - "$OUTPUT/countries.json" <<'PY'
import json
import pathlib
import sys

path = pathlib.Path(sys.argv[1])
catalog = json.loads(path.read_text())
codes = [entry["code"] for entry in catalog["countries"]]
expected = ["USA", "RUS", "CHN", "ESP", "GTM"]
if codes != expected:
    raise SystemExit(f"catálogo OCI inválido: {codes!r}")
print(f"[OK] Catálogo OCI descargado y validado: {path}")
PY
