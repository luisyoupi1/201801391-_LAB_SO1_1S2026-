# Resultados de pruebas ejecutadas

Fecha de verificación: 3 de agosto de 2026.
Entorno disponible: Windows, Go 1.21.5. El equipo de verificación no tiene Docker, Podman, Vagrant ni libvirt instalados.

## Código Go

| Verificación | Resultado |
|---|---|
| `gofmt` | Correcto |
| `go test ./...` | Correcto |
| `go vet ./...` | Correcto |
| Compilación de API1, API2 y API3 | Correcta |
| Cobertura de `internal/service` | 65.8% |

Las pruebas automatizadas cubren:

- Contrato y timestamp de `/health`.
- Comunicación cruzada exitosa.
- Respuesta remota con estado distinto de `UP`.
- API destino inalcanzable.
- Rechazo de una ruta con carnet incorrecto.
- Rechazo de métodos distintos de `GET`.

## Integración local

`scripts/local-integration-test.ps1` compiló y levantó los tres binarios simultáneamente. Resultados:

- 3 de 3 endpoints `/health` respondieron `UP`, VM y carnet correctos.
- 6 de 6 endpoints de comunicación cruzada respondieron `connection:true`.
- Los procesos temporales fueron detenidos al finalizar.

## Validación pendiente en el equipo KVM

La ejecución de Containerd, Podman, Docker, Zot, `push`/`pull` y las tres VMs no puede verificarse en este equipo porque esos componentes no están instalados. Debe completarse en el host Linux/KVM de entrega siguiendo `GUIA_INSTALACION.md` y documentarse con `EVIDENCIAS.md`. Esta distinción evita presentar como real una evidencia de infraestructura que todavía no se ha ejecutado.
