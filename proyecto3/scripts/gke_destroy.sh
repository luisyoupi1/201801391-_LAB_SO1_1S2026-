#!/usr/bin/env bash
set -euo pipefail

if [[ "${CONFIRM_GCP_DESTROY:-no}" != "yes" ]]; then
  echo "No se eliminó nada. Use CONFIRM_GCP_DESTROY=yes." >&2
  exit 2
fi
: "${GCP_PROJECT:?Defina GCP_PROJECT}"
: "${GCP_REGION:=us-central1}"
: "${GKE_CLUSTER:=so1-proyecto3}"
gcloud container clusters delete "$GKE_CLUSTER" --project "$GCP_PROJECT" --region "$GCP_REGION"
