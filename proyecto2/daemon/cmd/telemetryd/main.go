package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"proyecto2-so1-201801391/internal/config"
	"proyecto2-so1-201801391/internal/decision"
	"proyecto2-so1-201801391/internal/dockerctl"
	ebpfaudit "proyecto2-so1-201801391/internal/ebpf"
	metricserver "proyecto2-so1-201801391/internal/metrics"
	"proyecto2-so1-201801391/internal/model"
	"proyecto2-so1-201801391/internal/procfs"
	"proyecto2-so1-201801391/internal/valkeystore"
)

func main() {
	if err := run(); err != nil {
		log.Printf("ERROR: %v", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	manager := dockerctl.New(cfg.ProjectRoot, cfg.Carnet, cfg.ComposeFile)
	log.Printf("iniciando infraestructura para carnet %s", cfg.Carnet)
	if err := manager.StartInfrastructure(ctx); err != nil {
		return fmt.Errorf("start infrastructure: %w", err)
	}
	if err := manager.BuildWorkloads(ctx); err != nil {
		return fmt.Errorf("build workloads: %w", err)
	}
	if err := manager.LoadKernelModule(ctx); err != nil {
		return fmt.Errorf("load kernel module: %w", err)
	}
	if err := manager.InstallCron(ctx); err != nil {
		return fmt.Errorf("install cron: %w", err)
	}
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := manager.RemoveCron(cleanup); err != nil {
			log.Printf("no se pudo retirar cron: %v", err)
		}
	}()
	if err := manager.EnsureBaseline(ctx); err != nil {
		return fmt.Errorf("ensure initial container baseline: %w", err)
	}

	store := valkeystore.New(cfg.ValkeyAddr, cfg.ValkeyPassword, cfg.Carnet)
	defer store.Close()
	if err := store.WaitReady(ctx, 45*time.Second); err != nil {
		return err
	}

	registry := metricserver.New()
	go func() {
		if err := registry.Serve(ctx, cfg.MetricsAddr); err != nil {
			log.Printf("servidor de métricas: %v", err)
			stop()
		}
	}()

	auditor, err := ebpfaudit.Open(cfg.BPFObject)
	if err != nil && cfg.EBPFRequired {
		return err
	}
	if auditor == nil {
		log.Printf("ADVERTENCIA: eBPF no disponible: %v", err)
	} else {
		defer auditor.Close()
		auditor.Start(ctx, func(event model.KillEvent) {
			registry.ObserveEBPF()
			if err := store.RecordObservedKill(context.Background(), event); err != nil {
				log.Printf("guardar evento eBPF: %v", err)
			}
		})
	}

	runCycle := func() {
		if err := cycle(ctx, cfg, manager, store, registry, auditor); err != nil {
			log.Printf("ciclo de telemetría: %v", err)
		}
	}
	runCycle()
	ticker := time.NewTicker(cfg.LoopInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("finalizando servicio")
			return nil
		case <-ticker.C:
			runCycle()
		}
	}
}

func cycle(ctx context.Context, cfg config.Config, manager *dockerctl.Manager,
	store *valkeystore.Store, registry *metricserver.Registry, auditor *ebpfaudit.Auditor) error {
	snapshot, err := procfs.Read(cfg.ProcPath)
	if err != nil {
		return err
	}
	snapshot.Timestamp = time.Now().UTC()
	containers, err := manager.ListManaged(ctx)
	if err != nil {
		return err
	}
	containerMetrics := manager.Correlate(snapshot, containers)
	registry.Update(snapshot, containerMetrics)
	if err := store.RecordSnapshot(ctx, snapshot, containerMetrics); err != nil {
		return fmt.Errorf("record snapshot: %w", err)
	}

	victims := decision.SelectVictims(containerMetrics, cfg.MinLowContainers, cfg.MinHighContainers)
	for _, victim := range victims {
		if victim.Container.PID <= 0 || !victim.Container.Running {
			continue
		}
		var confirmation <-chan model.KillEvent
		if auditor != nil {
			confirmation = auditor.Expect(uint32(victim.Container.PID))
		}
		log.Printf("deteniendo %s perfil=%s score=%.2f", victim.Container.Name, victim.Container.Profile, victim.Score)
		if err := manager.Stop(ctx, victim.Container); err != nil {
			log.Printf("detener %s: %v", victim.Container.Name, err)
			continue
		}
		confirmed := false
		var event model.KillEvent
		if confirmation != nil {
			select {
			case observed, ok := <-confirmation:
				if ok {
					event, confirmed = observed, true
				}
			case <-time.After(cfg.DeleteConfirmTimeout):
				log.Printf("sin confirmación eBPF para PID %d", victim.Container.PID)
			case <-ctx.Done():
				return ctx.Err()
			}
		}
		if err := manager.Remove(ctx, victim.Container); err != nil {
			log.Printf("eliminar %s: %v", victim.Container.Name, err)
			continue
		}
		if confirmed {
			deletion := model.Deletion{Container: victim, Event: event, DeletedAt: time.Now().UTC()}
			if err := store.RecordDeletion(ctx, deletion); err != nil {
				log.Printf("guardar eliminación confirmada: %v", err)
			} else {
				registry.ObserveDeletion()
			}
		}
	}
	if err := manager.EnsureBaseline(ctx); err != nil {
		return fmt.Errorf("restore baseline: %w", err)
	}
	return nil
}
