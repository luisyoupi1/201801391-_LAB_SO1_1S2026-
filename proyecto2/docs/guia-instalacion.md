# Guía de instalación y uso

## 1. Requisitos

- Ubuntu 22.04 LTS o posterior, instalado directamente en la laptop.
- Arranque UEFI y almacenamiento AHCI.
- Usuario con `sudo`.
- Internet durante la instalación de dependencias e imágenes.
- Virtualización habilitada no es necesaria para Docker nativo.

## 2. Preparación

Abra una terminal dentro de `proyecto2`:

```bash
chmod +x scripts/*.sh
cp .env.example .env
sudo bash scripts/install_dependencies.sh
```

Si el instalador agregó el usuario al grupo `docker`, cierre sesión y vuelva a entrar.
Verifique:

```bash
docker run --rm hello-world
test -r /sys/kernel/btf/vmlinux && echo BTF_OK
```

## 3. Compilación

```bash
make workloads
make all CARNET=201801391
make test
```

Resultados esperados:

- `kernel/continfo.ko`;
- `daemon/internal/ebpf/kill_monitor.bpf.o`;
- `bin/telemetryd`;
- mensaje `Validación estructural correcta`.

## 4. Instalación como servicio

```bash
sudo make install CARNET=201801391
```

El instalador copia el proyecto a `/opt/so1-proyecto2-201801391`, instala el binario en
`/usr/local/bin` y activa `so1-telemetryd.service`.

```bash
sudo systemctl status so1-telemetryd
sudo journalctl -u so1-telemetryd -f
```

## 5. Pruebas funcionales

### Módulo

```bash
lsmod | grep continfo
cat /proc/continfo_pr2_so1_201801391 | jq '.memory, .processes[0:3]'
```

Prueba de descarga/carga:

```bash
sudo systemctl stop so1-telemetryd
sudo bash scripts/unload_module.sh
test ! -e /proc/continfo_pr2_so1_201801391 && echo PROC_ELIMINADO
sudo bash scripts/load_module.sh
sudo systemctl start so1-telemetryd
```

### Contenedores y cron

```bash
docker ps --filter label=so1.project=201801391
cat /etc/cron.d/so1-proyecto2-201801391
sudo bash scripts/generate_containers.sh
```

Después del siguiente ciclo deben permanecer al menos 3 perfiles `low` y 2 `high`.

### Valkey

```bash
docker exec so1-valkey-201801391 valkey-cli --scan --pattern 'so1:201801391:*'
docker exec so1-valkey-201801391 valkey-cli HGETALL so1:201801391:system:current
docker exec so1-valkey-201801391 valkey-cli LLEN so1:201801391:ebpf:events
```

### eBPF

Genere una señal controlada mientras el servicio está activo:

```bash
sleep 300 &
PID=$!
kill -TERM "$PID"
sudo journalctl -u so1-telemetryd -n 50 --no-pager
```

La métrica debe aumentar:

```bash
curl -s http://localhost:9105/metrics | grep so1_ebpf_kill_events_total
```

### Grafana

Abra <http://localhost:3000>.

- Usuario: `admin`
- Contraseña: `so1-201801391`
- Dashboard: **Sistemas Operativos 1 / SO1 - Contenedores 201801391**

Prometheus puede revisarse en <http://localhost:9090/targets>; el target debe estar `UP`.

## 6. Evidencias

```bash
sudo CARNET=201801391 PROJECT_ROOT=/opt/so1-proyecto2-201801391 \
  bash /opt/so1-proyecto2-201801391/scripts/collect_evidence.sh
```

Agregue manualmente capturas del dashboard y la terminal a la carpeta `evidencias`.

## 7. Solución de problemas

### `linux/headers` no encontrado

```bash
sudo apt update
sudo apt install linux-headers-$(uname -r)
```

### `/sys/kernel/btf/vmlinux` no existe

Actualice a un kernel genérico de Ubuntu con BTF y reinicie. Compruebe que
`linux-tools-$(uname -r)` esté instalado.

### `operation not permitted` al cargar eBPF

El servicio debe correr como root. Revise `dmesg` y confirme que Secure Boot no esté
bloqueando componentes no firmados.

### `invalid module format`

El módulo fue compilado para otro kernel:

```bash
make clean
make kernel
sudo systemctl restart so1-telemetryd
```

### Docker Compose no existe

La automatización acepta `docker compose` o `docker-compose`. Instale uno de los dos y
verifique su versión.

### Grafana no presenta datos

```bash
curl http://localhost:9105/metrics
curl http://localhost:9090/-/healthy
docker logs so1-prometheus-201801391
docker logs so1-grafana-201801391
```

## 8. Desinstalación

```bash
sudo make uninstall CARNET=201801391
```

La desinstalación conserva los volúmenes de Valkey, Prometheus y Grafana para evitar una
pérdida accidental de evidencias.

