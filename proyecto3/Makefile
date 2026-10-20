CARNET ?= 201801391
COMPOSE := docker compose --env-file .env -f docker-compose.yml

.PHONY: init build local-up local-down local-reset logs test validate smoke locust locust-headless zot-up zot-down oci-push oci-pull images-push k8s-render help

help:
	@echo "make init        Crea .env y certificados locales"
	@echo "make build       Construye todos los microservicios"
	@echo "make local-up    Levanta el flujo obligatorio local"
	@echo "make smoke       Prueba Locust->Rust->Go->gRPC->RabbitMQ->Valkey"
	@echo "make locust      Levanta la interfaz de Locust en :8089"
	@echo "make locust-headless Ejecuta una prueba de carga corta"
	@echo "make zot-up      Levanta Zot local con HTTPS"
	@echo "make oci-pull    Descarga y valida el catálogo OCI desde Zot"
	@echo "make test        Ejecuta pruebas Rust, Go y validaciones"
	@echo "make local-down  Apaga el proyecto y conserva volúmenes"

init:
	test -f .env || cp .env.example .env
	bash scripts/generate_local_certs.sh

build: init
	$(COMPOSE) build

local-up: build
	$(COMPOSE) up -d rabbitmq valkey go-api grpc-client grpc-server rabbit-writer consumer metrics-exporter rust-api prometheus grafana
	bash scripts/wait_local.sh

local-down:
	$(COMPOSE) --profile load --profile registry down

local-reset:
	$(COMPOSE) --profile load --profile registry down -v

logs:
	$(COMPOSE) logs -f --tail=100

test: init k8s-render
	docker build --target test -t so1-p3-rust-test ./api-rust
	docker build --target test -t so1-p3-go-test ./services-go
	python3 scripts/validate_project.py
	bash -n scripts/*.sh

validate:
	python3 scripts/validate_project.py

smoke:
	bash tests/smoke_local.sh

locust:
	$(COMPOSE) --profile load up -d locust
	@echo "Locust: http://localhost:$${LOCUST_PORT:-8089}"

locust-headless:
	bash scripts/run_locust_headless.sh

zot-up: init
	$(COMPOSE) --profile registry up -d zot
	bash scripts/wait_zot.sh

zot-down:
	$(COMPOSE) --profile registry stop zot

oci-push: zot-up
	bash scripts/push_oci_artifact.sh

oci-pull:
	bash scripts/pull_oci_artifact.sh

images-push: zot-up
	bash scripts/build_and_push_zot.sh

k8s-render:
	docker run --rm -v "$(CURDIR):/work" -w /work alpine/helm:3.18.4 template mum-operator operator/helm-charts/valkeyinstance >/dev/null
	docker run --rm -v "$(CURDIR):/work" -w /work alpine/helm:3.18.4 template mum-operator operator/helm-charts/grafanainstance >/dev/null
	docker run --rm -v "$(CURDIR):/work" -w /work registry.k8s.io/kubectl:v1.33.4 kustomize operator/config >/dev/null
	docker run --rm -v "$(CURDIR):/work" -w /work registry.k8s.io/kubectl:v1.33.4 kustomize k8s/gke >/dev/null
	docker run --rm -v "$(CURDIR):/work" -w /work registry.k8s.io/kubectl:v1.33.4 kustomize k8s/dapr >/dev/null
