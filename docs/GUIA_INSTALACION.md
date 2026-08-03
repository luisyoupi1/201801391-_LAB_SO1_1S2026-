# Guía de instalación y demostración

## 1. Requisitos del host

El despliegue evaluable necesita un host Linux con virtualización Intel VT-x o AMD-V habilitada en BIOS/UEFI. En Ubuntu o Debian:

```bash
sudo apt update
sudo apt install -y qemu-kvm libvirt-daemon-system libvirt-clients bridge-utils vagrant rsync
sudo usermod -aG libvirt,kvm "$USER"
```

Cierra la sesión y vuelve a entrar para aplicar los grupos. Después valida:

```bash
test -e /dev/kvm && echo "KVM disponible"
virsh list --all
vagrant plugin install vagrant-libvirt
```

Si el host es Windows, KVM no se ejecuta de forma nativa. Usa una instalación Linux física con virtualización anidada habilitada o una máquina Linux que el curso acepte. La prueba local de Go sí funciona en Windows, pero no sustituye la evidencia KVM.

## 2. Crear las tres VMs

Desde la raíz del proyecto:

```bash
vagrant up --provider=libvirt
vagrant status
virsh list --all
```

El `Vagrantfile` crea:

- `vm1`: `192.168.56.11`, 2 vCPU y 2 GiB RAM.
- `vm2`: `192.168.56.12`, 2 vCPU y 2 GiB RAM.
- `vm3`: `192.168.56.13`, 2 vCPU y 2 GiB RAM.

La carpeta del proyecto se sincroniza como `/proyecto` dentro de cada VM.

## 3. Instalar Zot en VM3

Ejecuta primero el registro:

```bash
vagrant ssh vm3 -c 'PROJECT_DIR=/proyecto bash /proyecto/scripts/vm3/setup-zot.sh'
curl http://192.168.56.13:5000/v2/
```

Una respuesta vacía `{}` con HTTP 200 confirma que Zot está listo.

## 4. Instalar Containerd y desplegar API1/API2 en VM1

```bash
vagrant ssh vm1 -c 'bash /proyecto/scripts/vm1/setup-containerd.sh'
vagrant ssh vm1 -c 'PROJECT_DIR=/proyecto bash /proyecto/scripts/vm1/deploy-apis.sh'
vagrant ssh vm1 -c 'sudo nerdctl ps'
```

El segundo script construye, publica, elimina, extrae y ejecuta las dos imágenes desde `192.168.56.13:5000`.

## 5. Instalar Podman y desplegar API3 en VM2

```bash
vagrant ssh vm2 -c 'bash /proyecto/scripts/vm2/setup-podman.sh'
vagrant ssh vm2 -c 'PROJECT_DIR=/proyecto bash /proyecto/scripts/vm2/deploy-api3.sh'
vagrant ssh vm2 -c 'sudo podman ps'
```

## 6. Probar el sistema completo

Desde el host Linux:

```bash
bash scripts/test-endpoints.sh
curl http://192.168.56.13:5000/v2/_catalog | jq .
```

El catálogo debe incluir `api1-201801391`, `api2-201801391` y `api3-201801391`.

Pruebas individuales útiles para capturas:

```bash
curl -s http://192.168.56.11:8081/health | jq .
curl -s http://192.168.56.11:8082/health | jq .
curl -s http://192.168.56.12:8083/health | jq .
curl -s http://192.168.56.11:8081/api1/201801391/call-api3 | jq .
curl -s http://192.168.56.12:8083/api3/201801391/call-api2 | jq .
```

## 7. Demostrar el manejo de errores

Detén API3 de forma temporal:

```bash
vagrant ssh vm2 -c 'sudo podman stop api3'
curl -s http://192.168.56.11:8081/api1/201801391/call-api3 | jq .
```

La salida debe mostrar `connection: false`. Restáurala inmediatamente:

```bash
vagrant ssh vm2 -c 'sudo podman start api3'
curl -s http://192.168.56.11:8081/api1/201801391/call-api3 | jq .
```

## 8. Prueba rápida sin VMs

En Windows con Go instalado:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\local-integration-test.ps1
```

En cualquier sistema con Docker Compose:

```bash
docker compose up --build -d
API1=http://127.0.0.1:8081 API2=http://127.0.0.1:8082 \
API3=http://127.0.0.1:8083 REGISTRY=http://127.0.0.1:5000 \
bash scripts/test-endpoints.sh
docker compose down
```

## 9. Publicar el repositorio solicitado

No incluyas `.vagrant`, binarios ni capturas con información sensible. Crea el repositorio privado con el nombre exacto del enunciado:

```bash
git init
git add .
git commit -m "Proyecto 1: APIs, KVM y registro Zot"
git branch -M main
git remote add origin URL_PRIVADA_DEL_REPOSITORIO
git push -u origin main
```

En GitHub, abre `Settings > Collaborators` y agrega `roldyoran`, `JoseLorenzana272` y `KINGR0X`. Verifica visualmente que las tres invitaciones aparezcan antes de entregar.

## 10. Apagar el laboratorio

Para apagar sin borrar las VMs:

```bash
vagrant halt
```

`vagrant destroy` elimina las VMs y su evidencia local; úsalo únicamente después de entregar y conservar tus capturas.
