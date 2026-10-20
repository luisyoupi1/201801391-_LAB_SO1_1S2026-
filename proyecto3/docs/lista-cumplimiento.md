# Lista de cumplimiento del enunciado

| Requisito | Implementación |
|---|---|
| Carpeta `proyecto3` | Raíz de este proyecto |
| Locust | `locust/locustfile.py` y Deployment K8s |
| Gateway API | `k8s/gke/gateway.yaml` y `httproute.yaml` |
| Ruta principal | `/grpc-201801391` |
| API REST Rust | `api-rust` |
| HPA 1–3, CPU 30 % | `k8s/base/rust-api.yaml` |
| Go API + gRPC Client, 2 contenedores | Deployment `go-entry` |
| gRPC Server + Writer, 2 contenedores | Deployment `go-processing` |
| RabbitMQ obligatorio | StatefulSet, writer y consumer |
| Persistencia Valkey | `ValkeyInstance`, StatefulSet y PVC |
| Grafana por CR | `GrafanaInstance`, chart y dashboard |
| Operator SDK Helm | `operator/` con dos watches y CRD |
| País asignado USA | Dashboard y consultas USA |
| Doce visualizaciones | Dashboard local y chart del operador |
| Zot HTTPS externo | Compose local y documentación de VM |
| OCI Artifact | ORAS publica y descarga `countries.json`; Locust consume la copia descargada |
| Manual Markdown | `docs/manual-tecnico.md` |
| Guía de instalación | `docs/guia-instalacion.md` |
| Dapr extra | Ruta, consumidor y manifiestos en `k8s/dapr` |
| Falco extra | Regla y values en `k8s/falco` |

Pendiente por decisión explícita de trabajar sólo localmente: crear recursos cobrables en GCP, instalar el sistema en GKE, publicar Zot en una VM externa y capturar resultados reales del entorno cloud.
