package ebpfaudit

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"
	"proyecto2-so1-201801391/internal/model"
)

type objects struct {
	TraceKill *ebpf.Program `ebpf:"trace_kill"`
	Events    *ebpf.Map     `ebpf:"events"`
}

type rawEvent struct {
	SourcePID   uint32
	TargetPID   uint32
	Signal      int32
	Padding     uint32
	TimestampNS uint64
	Command     [16]byte
}

type Auditor struct {
	objects objects
	link    link.Link
	reader  *ringbuf.Reader
	mu      sync.Mutex
	pending map[uint32][]chan model.KillEvent
	once    sync.Once
}

func Open(objectPath string) (*Auditor, error) {
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("remove memlock limit: %w", err)
	}
	spec, err := ebpf.LoadCollectionSpec(objectPath)
	if err != nil {
		return nil, fmt.Errorf("load BPF object %s: %w", objectPath, err)
	}
	auditor := &Auditor{pending: make(map[uint32][]chan model.KillEvent)}
	if err := spec.LoadAndAssign(&auditor.objects, nil); err != nil {
		return nil, fmt.Errorf("load BPF programs: %w", err)
	}
	auditor.link, err = link.Tracepoint("syscalls", "sys_enter_kill", auditor.objects.TraceKill, nil)
	if err != nil {
		auditor.closeObjects()
		return nil, fmt.Errorf("attach sys_enter_kill: %w", err)
	}
	auditor.reader, err = ringbuf.NewReader(auditor.objects.Events)
	if err != nil {
		_ = auditor.link.Close()
		auditor.closeObjects()
		return nil, fmt.Errorf("open eBPF ring buffer: %w", err)
	}
	return auditor, nil
}

func (a *Auditor) Expect(targetPID uint32) <-chan model.KillEvent {
	channel := make(chan model.KillEvent, 2)
	a.mu.Lock()
	a.pending[targetPID] = append(a.pending[targetPID], channel)
	a.mu.Unlock()
	return channel
}

func (a *Auditor) Start(ctx context.Context, observe func(model.KillEvent)) {
	go func() {
		<-ctx.Done()
		_ = a.reader.Close()
	}()
	go func() {
		for {
			record, err := a.reader.Read()
			if err != nil {
				if errors.Is(err, ringbuf.ErrClosed) {
					return
				}
				continue
			}
			var raw rawEvent
			if err := binary.Read(bytes.NewReader(record.RawSample), binary.LittleEndian, &raw); err != nil {
				continue
			}
			event := model.KillEvent{
				SourcePID: raw.SourcePID, TargetPID: raw.TargetPID, Signal: raw.Signal,
				TimestampNS: raw.TimestampNS,
				Command:     strings.TrimRight(string(raw.Command[:]), "\x00"),
				ObservedAt:  time.Now().UTC(),
			}
			if observe != nil {
				observe(event)
			}
			a.dispatch(event)
		}
	}()
}

func (a *Auditor) dispatch(event model.KillEvent) {
	a.mu.Lock()
	channels := a.pending[event.TargetPID]
	delete(a.pending, event.TargetPID)
	a.mu.Unlock()
	for _, channel := range channels {
		select {
		case channel <- event:
		default:
		}
		close(channel)
	}
}

func (a *Auditor) Close() error {
	var result error
	a.once.Do(func() {
		if a.reader != nil {
			result = a.reader.Close()
		}
		if a.link != nil {
			if err := a.link.Close(); result == nil {
				result = err
			}
		}
		a.closeObjects()
	})
	return result
}

func (a *Auditor) closeObjects() {
	if a.objects.TraceKill != nil {
		_ = a.objects.TraceKill.Close()
	}
	if a.objects.Events != nil {
		_ = a.objects.Events.Close()
	}
}
