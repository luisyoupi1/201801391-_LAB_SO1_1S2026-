#!/usr/bin/env python3
import json
import pathlib
import sys

ROOT = pathlib.Path(__file__).resolve().parents[1]
errors: list[str] = []


def require(path: str) -> pathlib.Path:
    target = ROOT / path
    if not target.exists():
        errors.append(f"falta {path}")
    return target


required = [
    "docker-compose.yml",
    "api-rust/src/main.rs",
    "services-go/proto/warreport.proto",
    "locust/locustfile.py",
    "grafana/dashboards/mum-dashboard.json",
    "operator/watches.yaml",
    "operator/config/crd/bases/so1.usac.edu.gt_valkeyinstances.yaml",
    "operator/config/crd/bases/so1.usac.edu.gt_grafanainstances.yaml",
    "k8s/gke/gateway.yaml",
    "k8s/gke/httproute.yaml",
    "docs/manual-tecnico.md",
    "docs/guia-instalacion.md",
]
for item in required:
    require(item)

dashboard_path = require("grafana/dashboards/mum-dashboard.json")
operator_dashboard_path = require("operator/helm-charts/grafanainstance/files/mum-dashboard.json")
if dashboard_path.exists():
    try:
        dashboard = json.loads(dashboard_path.read_text())
        titles = {p.get("title", "") for p in dashboard.get("panels", [])}
        expected = {
            "Máximo general de aviones", "Mínimo general de aviones",
            "Máximo general de barcos", "Mínimo general de barcos",
            "Top países por aviones", "Top países por barcos",
            "Moda de aviones", "Moda de barcos", "USA · aviones en aire",
            "USA · barcos en mar", "País asignado", "Reportes recibidos de USA",
        }
        missing = expected - titles
        if missing:
            errors.append("paneles faltantes: " + ", ".join(sorted(missing)))
    except (OSError, json.JSONDecodeError) as exc:
        errors.append(f"dashboard inválido: {exc}")

if dashboard_path.exists() and operator_dashboard_path.exists():
    try:
        local_dashboard = json.loads(dashboard_path.read_text())
        operator_dashboard = json.loads(operator_dashboard_path.read_text())
        if local_dashboard != operator_dashboard:
            errors.append("el dashboard local y el dashboard del Operator no coinciden")
    except (OSError, json.JSONDecodeError) as exc:
        errors.append(f"dashboard del Operator inválido: {exc}")

route = require("k8s/gke/httproute.yaml")
if route.exists() and "/grpc-201801391" not in route.read_text():
    errors.append("HTTPRoute no contiene /grpc-201801391")

dapr_component = require("k8s/dapr/components.yaml")
dapr_consumer = require("services-go/cmd/dapr-consumer/main.go")
if dapr_component.exists() and dapr_consumer.exists():
    subscription_route = "/dapr/war-reports"
    if subscription_route not in dapr_component.read_text() or subscription_route not in dapr_consumer.read_text():
        errors.append("la ruta de suscripción Dapr no coincide con el consumidor")

rbac = require("operator/config/rbac.yaml")
if rbac.exists() and ("coordination.k8s.io" not in rbac.read_text() or "leases" not in rbac.read_text()):
    errors.append("el Operator usa leader election pero no tiene RBAC para Leases")

oci_pull = require("scripts/pull_oci_artifact.sh")
locust = require("locust/locustfile.py")
if oci_pull.exists() and locust.exists() and "COUNTRIES_FILE" not in locust.read_text():
    errors.append("Locust no consume el catálogo descargado como artefacto OCI")

if errors:
    for error in errors:
        print(f"[FALLO] {error}")
    sys.exit(1)
print("[OK] Estructura, rutas, CRD y dashboard del Proyecto 3 validados")
