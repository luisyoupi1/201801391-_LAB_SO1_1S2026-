#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CERT_DIR="$ROOT/certs"
mkdir -p "$CERT_DIR"

if [[ -s "$CERT_DIR/zot.crt" && -s "$CERT_DIR/zot.key" ]]; then
  exit 0
fi

openssl req -x509 -newkey rsa:2048 -sha256 -nodes -days 365 \
  -keyout "$CERT_DIR/zot.key" \
  -out "$CERT_DIR/zot.crt" \
  -subj "/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,IP:127.0.0.1"
chmod 600 "$CERT_DIR/zot.key"
echo "Certificado local de Zot creado en $CERT_DIR"
