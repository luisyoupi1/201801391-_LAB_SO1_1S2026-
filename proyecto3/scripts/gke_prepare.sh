#!/usr/bin/env bash
set -euo pipefail

: "${GCP_PROJECT:?Defina GCP_PROJECT}"
: "${GCP_REGION:=us-central1}"
: "${GKE_CLUSTER:=so1-proyecto3}"

gcloud config set project "$GCP_PROJECT"
gcloud services enable container.googleapis.com compute.googleapis.com monitoring.googleapis.com
gcloud container clusters create "$GKE_CLUSTER" \
  --region "$GCP_REGION" \
  --machine-type n1-standard-2 \
  --num-nodes 1 \
  --gateway-api standard
gcloud container clusters get-credentials "$GKE_CLUSTER" --region "$GCP_REGION"
echo "[OK] Clúster preparado. Continúe con scripts/gke_deploy.sh"
