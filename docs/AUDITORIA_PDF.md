# Auditoría de cumplimiento del Proyecto 1

Revisión realizada el 19 de agosto de 2026 contra:

- `Proyecto1-SO1-2doSem.pdf` (SHA-256 `e5bef86d3cfa56781a7e91497249d7670ca52eb64af1fc67ab17d9e0a7cf8345`).
- `Guía_Calificación_Proyecto1SO1_Semestre2.pdf` (SHA-256 `73cb3ecc3d7c65959aab6ea7363a059eb5054b5de70e0e9f2dbb19b06f860366`).

## Estado comprobable en el repositorio

| Requisito | Implementación | Estado |
|---|---|---|
| Tres VMs sobre KVM | `Vagrantfile`, libvirt y tres VMs ejecutadas | Cumple; evidencia 01/02 |
| Containerd en VM1 | Containerd activo con API1/API2 e imágenes Zot | Cumple; evidencia 03 |
| Podman en VM2 | Podman rootless con API3 | Cumple; evidencia 04 |
| Docker y Zot en VM3 | Docker con Zot y catálogo de imágenes | Cumple; evidencia 05 |
| Tres APIs en Go | `api1`, `api2`, `api3` y paquete común `internal/service` | Cumple |
| Tres endpoints `/health` | JSON con `status`, `message`, `timestamp`, `VM` y `carnet` | Cumple |
| Seis llamadas cruzadas | Consulta HTTP real a `/health`, incluyendo manejo de error | Cumple |
| Imágenes `[API#-#CARNET]` | `api1-201801391`, `api2-201801391`, `api3-201801391` | Cumple; minúsculas por sintaxis OCI |
| Push/pull mediante Zot | Flujos ejecutados y tres repositorios disponibles | Cumple; evidencias 03, 04, 05 y 08 |
| Manual y guía | Manual técnico, manual de usuario y guía de instalación | Cumple |
| Pruebas y evidencias | Pruebas Go, integración, error y recuperación documentados | Cumple; evidencias 01 a 11 |
| Repositorio privado y colaboradores | Nombre y usuarios documentados | Pendiente de verificar en GitHub |

## Verificaciones completadas

1. Las tres VMs se ejecutaron bajo KVM.
2. VM1 usa Containerd, VM2 usa Podman y VM3 usa Docker.
3. Docker está ausente en VM1 y VM2.
4. Zot contiene las tres imágenes personalizadas.
5. Los tres `/health` y las seis llamadas cruzadas aprobaron.
6. El error `connection: false` y la recuperación posterior fueron comprobados.
7. Las pruebas Go y los Dockerfiles mínimos quedaron documentados.

Las evidencias se encuentran en [EVIDENCIAS.md](EVIDENCIAS.md). Solo queda la verificación externa de crear o renombrar el repositorio privado como `201801391_LAB_SO1_2S2026` y agregar a `JoseLorenzana272` y `KINGR0X` como colaboradores.
