CARNET ?= 201801391
GO ?= $(shell if test -x /usr/lib/go-1.24/bin/go; then echo /usr/lib/go-1.24/bin/go; else command -v go; fi)
ROOT := $(CURDIR)
BIN_DIR := $(ROOT)/bin
BPF_DIR := $(ROOT)/daemon/internal/ebpf

.PHONY: all kernel ebpf daemon test clean infra-up infra-down workloads install uninstall

all: kernel ebpf daemon

$(BIN_DIR):
	mkdir -p $(BIN_DIR)

kernel:
	$(MAKE) -C /lib/modules/$$(uname -r)/build M=$(ROOT)/kernel CARNET=$(CARNET) modules

$(BPF_DIR)/vmlinux.h:
	test -r /sys/kernel/btf/vmlinux
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > $@

ebpf: $(BPF_DIR)/vmlinux.h
	clang -O2 -g -Wall -Werror -target bpf -D__TARGET_ARCH_x86 \
		-I$(BPF_DIR) -c $(BPF_DIR)/kill_monitor.bpf.c \
		-o $(BPF_DIR)/kill_monitor.bpf.o

daemon: $(BIN_DIR)
	cd daemon && CGO_ENABLED=0 $(GO) build -buildvcs=false -trimpath -ldflags "-s -w" -o $(BIN_DIR)/telemetryd ./cmd/telemetryd

test:
	cd daemon && $(GO) test -race ./...
	cd daemon && $(GO) vet ./...
	python3 scripts/validate_project.py

infra-up:
	docker compose --env-file .env -f docker-compose.yml up -d valkey prometheus grafana

infra-down:
	docker compose --env-file .env -f docker-compose.yml down

workloads:
	docker compose --env-file .env -f docker-compose.yml --profile build build

install: all
	sudo CARNET=$(CARNET) PROJECT_ROOT=$(ROOT) bash scripts/install_service.sh

uninstall:
	sudo CARNET=$(CARNET) PROJECT_ROOT=$(ROOT) bash scripts/uninstall_service.sh

clean:
	-$(MAKE) -C /lib/modules/$$(uname -r)/build M=$(ROOT)/kernel clean
	rm -rf $(BIN_DIR) $(BPF_DIR)/kill_monitor.bpf.o
