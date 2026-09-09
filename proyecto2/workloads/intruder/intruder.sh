#!/bin/sh
set -u

deadline=$(( $(date +%s) + 240 ))
while [ "$(date +%s)" -lt "$deadline" ]; do
  for path in /etc/shadow /root/.ssh/id_rsa /host/etc/shadow; do
    if content="$(head -c 64 "$path" 2>&1)"; then
      echo "LECTURA_INESPERADA ruta=$path bytes=${#content}"
    else
      echo "ACCESO_DENEGADO ruta=$path detalle=$content"
    fi
  done
  sleep 10
done

