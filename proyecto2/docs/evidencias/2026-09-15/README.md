# Evidencias del Proyecto 2 — 15 de septiembre de 2026

Carnet: 201801391.

Las imágenes de comandos muestran sus salidas reales renderizadas en el navegador para facilitar la lectura; no son capturas de una ventana de terminal. Se incluyen los registros TXT completos y los comandos en `registros.json`. Grafana y Prometheus se capturaron directamente desde sus interfaces locales. La compilación y las pruebas se ejecutaron sobre una copia temporal del código para evitar modificar los artefactos instalados.

## Servicio Go y módulo cargado

Fecha: 2026-09-15T08:27:29.537975-06:00. Código de salida: 0.

![Servicio Go y módulo cargado](01-servicio.png)

[Registro completo](01-servicio.txt)

## Interfaz /proc: PID, memoria y CPU

Fecha: 2026-09-15T08:27:29.601616-06:00. Código de salida: 0.

![Interfaz /proc: PID, memoria y CPU](02-proc.png)

[Registro completo](02-proc.txt)

## Cron instalado: una ejecución cada minuto

Fecha: 2026-09-15T08:27:29.613206-06:00. Código de salida: 0.

![Cron instalado: una ejecución cada minuto](03-cron.png)

[Registro completo](03-cron.txt)

## Prueba real: cinco contenedores por ejecución

Fecha: 2026-09-15T08:27:32.051617-06:00. Código de salida: 0.

![Prueba real: cinco contenedores por ejecución](04-generador.png)

[Registro completo](04-generador.txt)

## Persistencia en Valkey

Fecha: 2026-09-15T08:27:32.599489-06:00. Código de salida: 0.

![Persistencia en Valkey](05-valkey.png)

[Registro completo](05-valkey.txt)

## Prueba controlada: SIGTERM capturado por eBPF

Fecha: 2026-09-15T08:27:34.782689-06:00. Código de salida: 0.

![Prueba controlada: SIGTERM capturado por eBPF](06-ebpf.png)

[Registro completo](06-ebpf.txt)

## Eliminaciones confirmadas y registradas

Fecha: 2026-09-15T08:27:34.948630-06:00. Código de salida: 0.

![Eliminaciones confirmadas y registradas](07-eliminaciones.png)

[Registro completo](07-eliminaciones.txt)

## Métricas del daemon

Fecha: 2026-09-15T08:27:35.097061-06:00. Código de salida: 0.

![Métricas del daemon](08-metricas.png)

[Registro completo](08-metricas.txt)

## Compilación real: kernel, eBPF y daemon Go

Fecha: 2026-09-15T08:29:55.297471-06:00. Código de salida: 0.

![Compilación real: kernel, eBPF y daemon Go](09-compilacion.png)

[Registro completo](09-compilacion.txt)

## Pruebas Go, detector de carreras y validación

Fecha: 2026-09-15T08:31:11.664571-06:00. Código de salida: 0.

![Pruebas Go, detector de carreras y validación](10-pruebas.png)

[Registro completo](10-pruebas.txt)

## Carga y descarga del módulo: /proc y servicio

Fecha: 2026-09-15T08:31:07-06:00. Código de salida: 0.

![Carga y descarga del módulo: /proc y servicio](14-carga-descarga.png)

[Registro completo](14-carga-descarga.txt)

## Prometheus: target UP

![Prometheus: target UP](11-prometheus.png)

## Grafana: estado y memoria

![Grafana: estado y memoria](12-grafana-estado.png)

## Grafana: consumo y eliminaciones

![Grafana: consumo y eliminaciones](13-grafana-actividad.png)

