# Proyecto 2 - Telemetría de contenedores

**Carnet:** 201801391  
**Curso:** Sistemas Operativos 1  
**Universidad:** Universidad de San Carlos de Guatemala

Implementación integral de telemetría y gestión de contenedores con:

- módulo de kernel en C que publica `/proc/continfo_pr2_so1_201801391`;
- sonda eBPF sobre `syscalls:sys_enter_kill` con eventos por ring buffer;
- daemon en Go que correlaciona procesos y contenedores, mantiene mínimos de carga y elimina excedentes;
- persistencia de métricas y eventos en Valkey;
- cron cada dos minutos para generar cinco contenedores aleatorios;
- Grafana y Prometheus provisionados automáticamente;
- imágenes personalizadas de memoria, CPU, bajo consumo e intruso.

## Vista general

### Dashboard de observabilidad

![Dashboard principal de Grafana](docs/evidencias/capturas/02-dashboard-grafana-metricas.png)

El dashboard concentra el estado de RAM, eventos eBPF, eliminaciones y evolución de las
métricas recolectadas por el daemon.

### Contenedores con mayor consumo

![Top 5 histórico de contenedores](docs/evidencias/capturas/03-dashboard-grafana-top5.png)

Los paneles históricos comparan los contenedores con mayor consumo de RAM y CPU, incluso
después de que hayan finalizado.

### Pipeline de métricas

![Target de Prometheus en estado UP](docs/evidencias/capturas/04-prometheus-target-up.png)

Prometheus consulta el exporter del daemon en el puerto 9105 y entrega las series a
Grafana. La ejecución completa está documentada en
[`docs/evidencias.md`](docs/evidencias.md).

## Inicio rápido

La ejecución debe hacerse en Ubuntu sobre el host real, no dentro de una máquina que no
exponga Docker, BTF y los encabezados del kernel.

```bash
cp .env.example .env
sudo ./scripts/install_dependencies.sh
make workloads
make all
make test
sudo make install
```

Estado del servicio:

```bash
sudo systemctl status so1-telemetryd
sudo journalctl -u so1-telemetryd -f
cat /proc/continfo_pr2_so1_201801391 | jq .
```

Grafana queda disponible en `http://localhost:3000` con usuario `admin` y contraseña
`so1-201801391`. El dashboard **SO1 - Contenedores 201801391** se provisiona al iniciar.

## Comandos útiles

```bash
make kernel              # compila el módulo C
make ebpf                # genera vmlinux.h y compila la sonda
make daemon              # compila el servicio Go
make workloads           # construye las imágenes de carga
make test                # pruebas unitarias y validación estructural
make infra-up            # Valkey, Prometheus y Grafana
sudo make install        # instala y activa el servicio
sudo make uninstall      # desinstala de forma limpia
```

La guía completa está en `docs/guia-instalacion.md`; el diseño, formato de datos y
decisiones técnicas se explican en `docs/manual-tecnico.md`.

## Validación rápida

```bash
systemctl is-active so1-telemetryd
systemctl is-enabled so1-telemetryd
lsmod | grep continfo
cat /proc/continfo_pr2_so1_201801391 | jq \
  '{carnet, memory, process_count:(.processes|length), sample:.processes[0:3]}'
sudo docker ps --filter label=so1.project=201801391
cat /etc/cron.d/so1-proyecto2-201801391
curl -fsS http://localhost:9105/metrics | grep '^so1_'
```

Se espera un servicio `active` y `enabled`, JSON válido en `/proc`, al menos tres
contenedores `low` y dos `high`, cron cada dos minutos, métricas `so1_*` y target de
Prometheus en estado `UP`.

Si Docker responde `permission denied`, vuelva a iniciar sesión para aplicar el grupo
`docker`, ejecute `newgrp docker` o use `sudo docker` durante la validación.

## Documentación y evidencias

- [Guía de instalación](docs/guia-instalacion.md)
- [Manual técnico](docs/manual-tecnico.md)
- [Guía para la defensa](docs/defensa.md)
- [Evidencias de ejecución](docs/evidencias.md)

La evidencia incluye las 18 capturas originales de Grafana, Prometheus, exporter,
systemd, procfs, cron, eBPF, generación de carga y logs del daemon.

## Seguridad

El contenedor intruso no monta el sistema de archivos del host. Intenta leer rutas
sensibles sin privilegios para producir evidencia de denegación sin exponer datos reales.
El daemon nunca elimina contenedores sin la etiqueta `so1.project=201801391`, y excluye
explícitamente la infraestructura de observabilidad.
