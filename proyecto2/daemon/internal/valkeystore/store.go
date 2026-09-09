package valkeystore

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"proyecto2-so1-201801391/internal/model"
)

type Store struct {
	client *redis.Client
	prefix string
}

func New(addr, password, carnet string) *Store {
	return &Store{
		client: redis.NewClient(&redis.Options{Addr: addr, Password: password}),
		prefix: "so1:" + carnet,
	}
}

func (s *Store) Close() error { return s.client.Close() }

func (s *Store) WaitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if err := s.client.Ping(ctx).Err(); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("Valkey not ready after %s", timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
}

func (s *Store) RecordSnapshot(ctx context.Context, snapshot model.Snapshot, containers []model.ContainerMetric) error {
	timestamp := snapshot.Timestamp.UnixMilli()
	pipe := s.client.TxPipeline()
	pipe.HSet(ctx, s.prefix+":system:current", map[string]interface{}{
		"timestamp_ms": timestamp,
		"ram_total_kb": snapshot.Memory.TotalKB,
		"ram_free_kb":  snapshot.Memory.FreeKB,
		"ram_used_kb":  snapshot.Memory.UsedKB,
	})
	pipe.Del(ctx, s.prefix+":containers:current")
	for name, value := range map[string]uint64{
		"ram_total_kb": snapshot.Memory.TotalKB,
		"ram_free_kb":  snapshot.Memory.FreeKB,
		"ram_used_kb":  snapshot.Memory.UsedKB,
	} {
		member := strconv.FormatInt(timestamp, 10) + ":" + strconv.FormatUint(value, 10)
		pipe.ZAdd(ctx, s.prefix+":series:"+name, redis.Z{Score: float64(timestamp), Member: member})
	}
	for _, container := range containers {
		payload, _ := json.Marshal(container)
		pipe.HSet(ctx, s.prefix+":containers:current", container.Container.ID, payload)
		pipe.HSet(ctx, s.prefix+":containers:history", container.Container.ID, payload)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *Store) RecordObservedKill(ctx context.Context, event model.KillEvent) error {
	payload, _ := json.Marshal(event)
	return s.client.LPush(ctx, s.prefix+":ebpf:events", payload).Err()
}

func (s *Store) RecordDeletion(ctx context.Context, deletion model.Deletion) error {
	payload, _ := json.Marshal(deletion)
	timestamp := deletion.DeletedAt.UnixMilli()
	pipe := s.client.TxPipeline()
	pipe.Incr(ctx, s.prefix+":deleted:total")
	pipe.ZAdd(ctx, s.prefix+":series:deletions", redis.Z{Score: float64(timestamp), Member: payload})
	pipe.HDel(ctx, s.prefix+":containers:current", deletion.Container.Container.ID)
	_, err := pipe.Exec(ctx)
	return err
}
