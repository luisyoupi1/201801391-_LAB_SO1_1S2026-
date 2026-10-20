# Guía de pruebas y evidencias

Guarde las capturas en `docs/evidencias/` y reemplace cada casilla cuando se ejecute en GKE.

## Evidencia local

- [x] `make test` finaliza sin errores.
- [x] `make smoke` muestra el flujo completo.
- [x] Locust envía solicitudes a `/grpc-201801391` sin fallos.
- [x] RabbitMQ transporta los mensajes de la prueba.
- [x] Valkey contiene `mum:reports:total`.
- [x] Prometheus devuelve `mum_reports_total` y ambos targets están activos.
- [x] Grafana tiene el dashboard de doce paneles provisionado.
- [x] Zot responde por HTTPS y contiene las imágenes más `so1/countries:v1`.

Resultado local validado el 14 de octubre de 2026: 10 usuarios, 30 segundos, 1,119 solicitudes, 37.77 solicitudes/s, p95 de 7 ms y 0 fallos. El reporte HTML queda en `runtime/locust/report.html`. Este resultado comprueba la integración local; no sustituye la comparación requerida en GKE.

Capturas locales incluidas:

- [Dashboard de Grafana](evidencias/grafana-local.png)
- [Reporte de Locust](evidencias/locust-local.png)

Comandos de apoyo:

```bash
docker compose ps
docker compose exec valkey valkey-cli GET mum:reports:total
docker compose exec valkey valkey-cli HGETALL mum:country:counts
curl -s http://localhost:9106/metrics | grep '^mum_'
curl -s 'http://localhost:9090/api/v1/query?query=mum_reports_total'
curl -k https://localhost:5000/v2/_catalog
```

## Evidencia GKE pendiente

- [ ] Nodos GKE con tipo N1.
- [ ] `Gateway` y `HTTPRoute` aceptados, con IP pública.
- [ ] Ruta `/grpc-201801391` funcional desde Locust.
- [ ] HPA pasa de una a más réplicas durante carga.
- [ ] `ValkeyInstance` reconcilia StatefulSet, Service y PVC.
- [ ] `GrafanaInstance` reconcilia StatefulSet, Service, PVC y dashboard.
- [ ] Imágenes descargadas desde Zot de la VM.
- [ ] Prueba con una réplica de `go-processing`.
- [ ] Prueba con dos réplicas de `go-processing`.
- [ ] Resultado y conclusión de latencia p95, RPS y fallos.
- [ ] Si se presenta Dapr: ruta, sidecars, componente y suscripción.
- [ ] Si se presenta Falco: regla y alerta en logs.

Tabla para resultados reales:

| Escenario | Usuarios | Réplicas Go | RPS | p95 ms | Fallos |
|---|---:|---:|---:|---:|---:|
| Base | Pendiente | 1 | Pendiente | Pendiente | Pendiente |
| Escalado | Pendiente | 2 | Pendiente | Pendiente | Pendiente |

No se inventan resultados: esta tabla debe llenarse durante la ejecución final en GKE.
