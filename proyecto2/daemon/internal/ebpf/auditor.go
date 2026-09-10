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
	TraceSignal *ebpf.Program `ebpf:"trace_signal"`
	TraceKill   *ebpf.Program `ebpf:"trace_kill"`
	Events      *ebpf.Map     `ebpf:"events"`
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
	objects    objects
	link       link.Link
	signalLink link.Link
	reader     *ringbuf.Reader
	mu         sync.Mutex
	pending    map[uint32][]chan model.KillEvent
	once       sync.Once
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
	auditor.signalLink, err = link.AttachRawTracepoint(link.RawTracepointOptions{Name: "signal_generate", Program: auditor.objects.TraceSignal})
	if err != nil {
		_ = auditor.link.Close()
		auditor.closeObjects()
		return nil, fmt.Errorf("attach signal_generate: %w", err)
	}
	auditor.reader, err = ringbuf.NewReader(auditor.objects.Events)
	if err != nil {
		_ = auditor.link.Close()
		auditor.closeObjects()
		_ = auditor.signalLink.Close()
		return nil, fmt.Errorf("open eBPF ring buffer: %w", err)
	}
	return auditor, nil
}

func (a *Auditor) Expect(targetPID uint32) (<-chan model.KillEvent, func()) {
	channel := make(chan model.KillEvent, 2)
	a.mu.Lock()
	a.pending[targetPID] = append(a.pending[targetPID], channel)
	a.mu.Unlock()
	return channel, func() {
		a.mu.Lock()
		defer a.mu.Unlock()
		channels := a.pending[targetPID]
		for i, candidate := range channels {
			if candidate == channel {
				channels = append(channels[:i], channels[i+1:]...)
				if len(channels) == 0 {
					delete(a.pending, targetPID)
				} else {
					a.pending[targetPID] = channels
				}
				close(channel)
				break
			}
		}
	}
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
			event.Origin = "sys_enter_kill"
			if raw.Padding == 1 {
				event.Origin = "signal_generate"
			}
			// Deliver confirmations before persistence, which may block.
			a.dispatch(event)
			if observe != nil {
				observe(event)
			}
		}
	}()
}

func (a *Auditor) dispatch(event model.KillEvent) {
	if event.Origin != "signal_generate" || (event.Signal != 9 && event.Signal != 15) {
		return
	}
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
		if a.signalLink != nil {
			_ = a.signalLink.Close()
		}
		a.mu.Lock()
		for pid, channels := range a.pending {
			for _, channel := range channels {
				close(channel)
			}
			delete(a.pending, pid)
		}
		a.mu.Unlock()
		a.closeObjects()
	})
	return result
}

func (a *Auditor) closeObjects() {
	if a.objects.TraceSignal != nil {
		_ = a.objects.TraceSignal.Close()
	}
	if a.objects.TraceKill != nil {
		_ = a.objects.TraceKill.Close()
	}
	if a.objects.Events != nil {
		_ = a.objects.Events.Close()
	}
}
