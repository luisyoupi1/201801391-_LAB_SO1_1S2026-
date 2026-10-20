# Despliegue en GCP preparado para después

Estos pasos crean recursos cobrables. No se han ejecutado durante la preparación local.

## 1. Crear Zot fuera del clúster

1. Crear una VM Compute Engine N1 con IP estática.
2. Apuntar un dominio a la IP y emitir un certificado TLS válido.
3. Copiar `zot/config.json`, adaptar las rutas del certificado y ejecutar Zot.
4. Permitir TCP 443 o el puerto configurado únicamente desde los orígenes necesarios.
5. Definir `ZOT_REGISTRY=dominio:puerto`, iniciar sesión si se agrega autenticación y ejecutar `make images-push` y `make oci-push`.
6. Construir y publicar también el operador:

```bash
docker build -t "$ZOT_REGISTRY/so1/mum-operator:v1" operator
docker push "$ZOT_REGISTRY/so1/mum-operator:v1"
```

## 2. Crear GKE

```bash
export GCP_PROJECT=ID_DEL_PROYECTO
export GCP_REGION=us-central1
export GKE_CLUSTER=so1-proyecto3
bash scripts/gke_prepare.sh
```

El script solicita nodos `n1-standard-2` y activa Gateway API.

## 3. Cambiar secretos

Antes de aplicar, cambie los valores `CAMBIAR_ANTES_DE_GKE` de `k8s/base/config.yaml`. Para una entrega real conviene usar Secret Manager o un Secret creado fuera de Git.

## 4. Desplegar

```bash
export IMAGE_REGISTRY=zot.midominio.com
CONFIRM_GCP_DEPLOY=yes bash scripts/gke_deploy.sh
```

El script reemplaza `zot.example.com:5000` al renderizar; no modifica los archivos fuente. Luego espere la dirección del Gateway:

```bash
kubectl get gateway mum-gateway -n so1-proyecto3 -w
kubectl describe httproute rust-api -n so1-proyecto3
```

## 5. Acceder a herramientas internas

```bash
kubectl port-forward -n so1-proyecto3 svc/grafana 3000:3000
kubectl port-forward -n so1-proyecto3 svc/prometheus 9090:9090
kubectl port-forward -n so1-proyecto3 svc/locust 8089:8089
kubectl port-forward -n so1-proyecto3 svc/rabbitmq 15672:15672
```

## 6. Probar réplicas

```bash
kubectl scale deployment/go-processing -n so1-proyecto3 --replicas=1
# ejecutar el mismo escenario Locust y guardar resultados
kubectl scale deployment/go-processing -n so1-proyecto3 --replicas=2
# repetir exactamente usuarios, duración y tasa
```

## 7. Extras

Dapr requiere instalar su control plane y aplicar `k8s/dapr`. Falco se instala con `scripts/install_falco_optional.sh`. Ambos deben ejecutarse únicamente cuando el flujo obligatorio ya funcione.

## 8. Eliminar el clúster

```bash
CONFIRM_GCP_DESTROY=yes bash scripts/gke_destroy.sh
```
