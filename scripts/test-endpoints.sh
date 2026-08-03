#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-201801391}"
API1="${API1:-http://192.168.56.11:8081}"
API2="${API2:-http://192.168.56.11:8082}"
API3="${API3:-http://192.168.56.12:8083}"
REGISTRY="${REGISTRY:-http://192.168.56.13:5000}"

require_value() {
  local url="$1"
  local expression="$2"
  echo "GET ${url}"
  curl --fail --silent --show-error "$url" | tee /tmp/so1-response.json | jq .
  jq --exit-status "$expression" /tmp/so1-response.json >/dev/null
}

require_value "${API1}/health" ".status == \"UP\" and .VM == \"VM1\" and .carnet == \"${CARNET}\""
require_value "${API2}/health" ".status == \"UP\" and .VM == \"VM1\" and .carnet == \"${CARNET}\""
require_value "${API3}/health" ".status == \"UP\" and .VM == \"VM2\" and .carnet == \"${CARNET}\""

require_value "${API1}/api1/${CARNET}/call-api2" ".apiname == \"API2\" and .connection == true and .carnet == \"${CARNET}\""
require_value "${API1}/api1/${CARNET}/call-api3" ".apiname == \"API3\" and .connection == true and .carnet == \"${CARNET}\""
require_value "${API2}/api2/${CARNET}/call-api1" ".apiname == \"API1\" and .connection == true and .carnet == \"${CARNET}\""
require_value "${API2}/api2/${CARNET}/call-api3" ".apiname == \"API3\" and .connection == true and .carnet == \"${CARNET}\""
require_value "${API3}/api3/${CARNET}/call-api1" ".apiname == \"API1\" and .connection == true and .carnet == \"${CARNET}\""
require_value "${API3}/api3/${CARNET}/call-api2" ".apiname == \"API2\" and .connection == true and .carnet == \"${CARNET}\""

echo "GET ${REGISTRY}/v2/_catalog"
curl --fail --silent --show-error "${REGISTRY}/v2/_catalog" | jq .
echo "Todos los contratos REST y el registro respondieron correctamente."
