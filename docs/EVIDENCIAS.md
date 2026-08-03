# Plan de capturas y evidencia funcional

Guarda las imágenes en `evidencias/capturas/` y reemplaza cada casilla con el nombre real. No inventes resultados: toma las capturas después de ejecutar el laboratorio KVM.

| No. | Evidencia | Comando o pantalla | Archivo sugerido |
|---:|---|---|---|
| 1 | KVM habilitado | `virsh list --all` y `/dev/kvm` | `01-kvm-vms.png` |
| 2 | Tres VMs activas | `vagrant status` | `02-vagrant-status.png` |
| 3 | Containerd en VM1 | `sudo nerdctl version && sudo nerdctl ps` | `03-containerd-vm1.png` |
| 4 | Podman en VM2 | `sudo podman version && sudo podman ps` | `04-podman-vm2.png` |
| 5 | Docker y Zot en VM3 | `sudo docker ps` y `curl localhost:5000/v2/` | `05-zot-vm3.png` |
| 6 | Imágenes en Zot | `curl 192.168.56.13:5000/v2/_catalog` | `06-zot-catalogo.png` |
| 7 | Tres `/health` | Tres comandos `curl .../health` | `07-health.png` |
| 8 | Seis llamadas cruzadas | `bash scripts/test-endpoints.sh` | `08-cross-api.png` |
| 9 | Error controlado | API3 detenida y `connection:false` | `09-error-controlado.png` |
| 10 | Recuperación | API3 iniciada y `connection:true` | `10-recuperacion.png` |
| 11 | Pruebas Go | `go test ./...` | `11-go-test.png` |
| 12 | Repositorio privado | Estructura y colaboradores | `12-github.png` |

## Registro de resultados

Completa esta tabla durante la demostración:

| Prueba | Resultado esperado | Resultado observado | Cumple |
|---|---|---|---|
| API1 `/health` | `UP`, VM1, carnet correcto | Pendiente | Pendiente |
| API2 `/health` | `UP`, VM1, carnet correcto | Pendiente | Pendiente |
| API3 `/health` | `UP`, VM2, carnet correcto | Pendiente | Pendiente |
| API1 a API2/API3 | `connection:true` | Pendiente | Pendiente |
| API2 a API1/API3 | `connection:true` | Pendiente | Pendiente |
| API3 a API1/API2 | `connection:true` | Pendiente | Pendiente |
| API destino apagada | `connection:false` | Pendiente | Pendiente |
| Catálogo Zot | tres repositorios | Pendiente | Pendiente |

## Evidencia de push/pull

Además de las capturas, conserva la salida de:

```bash
vagrant ssh vm1 -c 'sudo nerdctl images'
vagrant ssh vm2 -c 'sudo podman images'
curl -s http://192.168.56.13:5000/v2/_catalog | jq .
```
