#!/usr/bin/env python3
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
CARNET = "201801391"

required = [
    "kernel/continfo.c",
    "kernel/Makefile",
    "daemon/cmd/telemetryd/main.go",
    "daemon/internal/ebpf/kill_monitor.bpf.c",
    "docker-compose.yml",
    "grafana/dashboards/so1-containers.json",
    "scripts/generate_containers.sh",
    "docs/manual-tecnico.md",
    "docs/guia-instalacion.md",
]
missing = [name for name in required if not (ROOT / name).is_file()]
if missing:
    raise SystemExit(f"Faltan archivos: {missing}")

dashboard = json.loads((ROOT / "grafana/dashboards/so1-containers.json").read_text(encoding="utf-8"))
panels = dashboard.get("panels", [])
ids = [panel["id"] for panel in panels]
if len(ids) != len(set(ids)):
    raise SystemExit("IDs de panel duplicados")
required_metrics = {
    "so1_ram_total_bytes", "so1_ram_free_bytes", "so1_ram_used_bytes",
    "so1_containers_deleted_total", "so1_ebpf_kill_events_total",
    "so1_container_memory_percent", "so1_container_cpu_percent",
    "so1_last_success_timestamp_seconds",
}
expressions = "\n".join(target.get("expr", "") for panel in panels for target in panel.get("targets", []))
for metric in required_metrics:
    if metric not in expressions:
        raise SystemExit(f"Falta métrica en dashboard: {metric}")
for index, panel in enumerate(panels):
    box = panel["gridPos"]
    if box["x"] < 0 or box["w"] <= 0 or box["h"] <= 0 or box["x"] + box["w"] > 24:
        raise SystemExit(f"Panel fuera de cuadrícula: {panel['id']}")
    for other in panels[:index]:
        prior = other["gridPos"]
        if (box["x"] < prior["x"] + prior["w"] and prior["x"] < box["x"] + box["w"]
            and box["y"] < prior["y"] + prior["h"] and prior["y"] < box["y"] + box["h"]):
            raise SystemExit(f"Paneles superpuestos: {panel['id']} y {other['id']}")

kernel = (ROOT / "kernel/continfo.c").read_text(encoding="utf-8")
for token in ("task_struct", "get_task_mm", "task_cputime_adjusted", "PROC_NAME"):
    if token not in kernel:
        raise SystemExit(f"El módulo no contiene {token}")

all_text = "\n".join(path.read_text(encoding="utf-8", errors="ignore") for path in ROOT.rglob("*") if path.is_file())
if CARNET not in all_text:
    raise SystemExit("El carnet no está configurado")

print("Validación estructural correcta para carnet", CARNET)

