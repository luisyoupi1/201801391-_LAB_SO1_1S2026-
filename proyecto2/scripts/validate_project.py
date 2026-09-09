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
titles = {panel.get("title") for panel in dashboard.get("panels", [])}
expected = {
    "Total de RAM",
    "RAM usada",
    "Memoria libre",
    "Total contenedores eliminados",
    "Uso de RAM a lo largo del tiempo",
    "Contenedores eliminados a lo largo del tiempo",
    "Top 5 contenedores por RAM (histórico)",
    "Top 5 contenedores por CPU (histórico)",
    "Eventos eBPF sys_kill",
}
if titles != expected:
    raise SystemExit(f"Paneles inesperados. Faltan={expected-titles}; sobran={titles-expected}")

kernel = (ROOT / "kernel/continfo.c").read_text(encoding="utf-8")
for token in ("task_struct", "get_task_mm", "task_cputime_adjusted", "PROC_NAME"):
    if token not in kernel:
        raise SystemExit(f"El módulo no contiene {token}")

all_text = "\n".join(path.read_text(encoding="utf-8", errors="ignore") for path in ROOT.rglob("*") if path.is_file())
if CARNET not in all_text:
    raise SystemExit("El carnet no está configurado")

print("Validación estructural correcta para carnet", CARNET)

