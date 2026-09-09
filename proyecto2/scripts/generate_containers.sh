#!/usr/bin/env bash
set -euo pipefail

CARNET="${CARNET:-201801391}"
PROJECT_LABEL="so1.project=${CARNET}"

running_count() {
  docker ps --filter "label=${PROJECT_LABEL}" --filter "label=so1.profile=$1" -q | wc -l
}

launch() {
  local profile="$1"
  local image resources
  case "$profile" in
    low)
      image="so1-${CARNET}-low:latest"
      resources=(--memory 64m --cpus 0.25)
      ;;
    high-memory)
      image="so1-${CARNET}-memory:latest"
      profile="high"
      resources=(--memory 256m --cpus 0.75)
      ;;
    high-cpu)
      image="so1-${CARNET}-cpu:latest"
      profile="high"
      resources=(--memory 96m --cpus 1.5)
      ;;
    intruder)
      image="so1-${CARNET}-intruder:latest"
      resources=(--memory 64m --cpus 0.25 --security-opt no-new-privileges)
      ;;
    *)
      echo "perfil desconocido: $profile" >&2
      return 1
      ;;
  esac
  local name="so1-${CARNET}-${profile}-$(date +%s)-${RANDOM}"
  docker run --rm -d --name "$name" \
    --label "$PROJECT_LABEL" --label "so1.profile=${profile}" \
    "${resources[@]}" "$image" >/dev/null
  echo "creado $name ($image)"
}

ensure_baseline() {
  local low high
  low="$(running_count low)"
  high="$(running_count high)"
  while [ "$low" -lt 3 ]; do
    launch low
    low=$((low + 1))
  done
  while [ "$high" -lt 2 ]; do
    if [ $((RANDOM % 2)) -eq 0 ]; then launch high-memory; else launch high-cpu; fi
    high=$((high + 1))
  done
}

if [ "${1:-}" = "--baseline" ]; then
  ensure_baseline
  exit 0
fi

profiles=(low high-memory high-cpu intruder)
for _ in 1 2 3 4 5; do
  launch "${profiles[$((RANDOM % ${#profiles[@]}))]}"
done
ensure_baseline

