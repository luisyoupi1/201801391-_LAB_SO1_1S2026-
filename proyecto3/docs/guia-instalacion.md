# Guía de instalación

## 1. Requisitos locales

- Linux con Docker Engine y Docker Compose v2.
- `make`, `curl`, `openssl` y Python 3.
- Al menos 6 GB de RAM disponibles y 10 GB de disco.

No es necesario instalar Go, Rust, Protobuf, Locust, Prometheus o Grafana en el host. Las compilaciones y servicios se ejecutan en contenedores.

## 2. Preparar el proyecto

```bash
git clone https://github.com/luisyoupi1/201801391-_LAB_SO1_1S2026-.git
cd 201801391-_LAB_SO1_1S2026-/proyecto3
make init
```

`make init` crea `.env` desde `.env.example` y genera el certificado local de Zot. Para cambiar puertos o contraseñas, edite `.env`.

## 3. Ejecutar el flujo obligatorio

```bash
make local-up
make smoke
```

El segundo comando envía un reporte y verifica su recorrido completo hasta Valkey, Prometheus y Grafana.

## 4. Ejecutar Locust

Interfaz web:

```bash
make locust
```

Abra <http://localhost:8089>, use `http://rust-api:8080` como host, defina usuarios y tasa de aparición, y comience la prueba.

Prueba automática de 30 segundos:

```bash
make locust-headless
```

Se puede ajustar con `USERS=50 SPAWN_RATE=5 DURATION=2m make locust-headless`.

## 5. Probar Zot y el artefacto OCI

```bash
make zot-up
make oci-push
curl -k https://localhost:5000/v2/_catalog
```

El artefacto `artifacts/countries.json` se publica como `so1/countries:v1`, se descarga desde Zot a `runtime/oci/countries.json` y se valida. Locust lee esa copia descargada para obtener los países permitidos. También puede repetirse solamente la descarga con `make oci-pull`.

## 6. Ejecutar todas las verificaciones

```bash
make test
make validate
make smoke
```

## 7. Apagar

```bash
make local-down
```

Para borrar también los datos persistentes:

```bash
make local-reset
```

## 8. Herramientas para el despliegue futuro

Antes de usar los scripts de GKE se necesitan `gcloud`, `kubectl` y `helm`, una cuenta de GCP con facturación, un dominio con TLS válido para Zot y las imágenes publicadas. Consulte [despliegue-gcp-pendiente.md](despliegue-gcp-pendiente.md). Los scripts de despliegue y eliminación requieren variables explícitas para evitar una ejecución accidental.
