# Proyecto 3 — M.U.M. con Kubernetes

Implementación del Proyecto 3 de Sistemas Operativos 1 para el carné **201801391**. El flujo obligatorio recibe reportes militares, los transporta mediante gRPC y RabbitMQ, los almacena en Valkey y los presenta en Grafana.

## Ejecución local

```bash
cd proyecto3
make local-up
make smoke
make locust
```

Para probar también el archivo de entrada distribuido como artefacto OCI:

```bash
make oci-push
make locust-headless
```

`make oci-push` publica y vuelve a descargar `countries.json` desde Zot. Locust usa esa copia descargada desde `runtime/oci/countries.json`.

Servicios locales:

| Servicio | URL | Credenciales |
|---|---|---|
| API Rust | <http://localhost:8080/health> | — |
| Grafana | <http://localhost:3000> | `admin` / `so1-201801391` |
| Prometheus | <http://localhost:9090> | — |
| RabbitMQ | <http://localhost:15672> | `so1` / `so1-local-201801391` |
| Locust | <http://localhost:8089> | — |
| Zot | <https://localhost:5000/v2/> | certificado local autofirmado |

La ruta principal es `POST /grpc-201801391`. La ruta extra preparada para Dapr es `POST /dapr-201801391`.

## Documentación

- [Manual técnico](docs/manual-tecnico.md)
- [Guía de instalación](docs/guia-instalacion.md)
- [Guía de pruebas y evidencias](docs/evidencias.md)
- [Lista de cumplimiento](docs/lista-cumplimiento.md)
- [Preparación para GCP](docs/despliegue-gcp-pendiente.md)

El despliegue a GCP, la publicación externa en Zot y la instalación opcional de Dapr/Falco quedan preparados, pero sus scripts no se ejecutan automáticamente.
