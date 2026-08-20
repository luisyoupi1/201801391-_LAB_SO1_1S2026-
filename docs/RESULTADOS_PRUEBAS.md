# Resultados de pruebas ejecutadas

Fecha de verificación: 19 de agosto de 2026 (hora de Guatemala; 20 de agosto en las VMs configuradas en UTC).

## Código e imágenes

| Verificación | Resultado |
|---|---|
| `gofmt` | Correcto |
| `go test ./...` | Correcto |
| `go vet ./...` | Correcto |
| Compilación de API1, API2 y API3 | Correcta |
| Construcción de las tres imágenes OCI | Correcta |
| Tamaño por imagen final | 6.337 MB en Containerd (aprox. 6.0 MiB) |
| `git diff --check` y sintaxis Bash | Correctos |

Las pruebas Go cubren el contrato de `/health`, comunicación exitosa, estado remoto distinto de `UP`, destino inalcanzable, carnet incorrecto y métodos distintos de `GET`.

## Laboratorio KVM real

Las tres VMs se ejecutaron con KVM desde discos `qcow2` almacenados en el USB `SO1_VMS`:

| VM | IP | Runtime | Servicio | Resultado |
|---|---|---|---|---|
| `so1-vm1` | `192.168.56.11` | Containerd 2.2.1 + nerdctl | API1 y API2 | Correcto |
| `so1-vm2` | `192.168.56.12` | Podman 4.9.3 rootless | API3 | Correcto |
| `so1-vm3` | `192.168.56.13` | Docker 29.1.3 | Zot | Correcto |

`systemd-detect-virt` devolvió `kvm` en las tres VMs. Docker no está instalado en VM1 ni VM2. Se habilitó `linger` para que API3 rootless permanezca en ejecución después de cerrar SSH.

## Integración y registro

La ejecución de `scripts/test-endpoints.sh` desde VM1 obtuvo:

- 3 de 3 endpoints `/health` con `status: "UP"`, VM y carnet correctos.
- 6 de 6 endpoints de comunicación cruzada con `connection: true`.
- Catálogo Zot con `api1-201801391`, `api2-201801391` y `api3-201801391`.
- Push, eliminación local y pull exitosos para las imágenes servidas por Zot.

También se detuvo API3 temporalmente. API1 respondió con `connection: false`, carnet correcto y `ERROR: The API3 located on the VM2 is not working`. Después de reiniciar API3, su `/health` volvió a `UP` y la pasada completa volvió a aprobar.

Las salidas originales y capturas de cada criterio se encuentran en [EVIDENCIAS.md](EVIDENCIAS.md).

![Resultado final de integración y catálogo Zot](evidencias/capturas/08-prueba-final-catalogo.jpg)

## Uso de almacenamiento

El laboratorio ocupa 4.7 GB en el USB de 57 GB utilizables y deja aproximadamente 53 GB libres. El directorio de las tres VMs ocupa 4.1 GB; la imagen base compartida ocupa 596 MB.
