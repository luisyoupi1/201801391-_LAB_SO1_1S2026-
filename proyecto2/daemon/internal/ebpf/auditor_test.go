package ebpfaudit

import (
	"proyecto2-so1-201801391/internal/model"
	"sync"
	"testing"
)

func TestConfirmOnlyAcceptedTermination(t *testing.T) {
	a := &Auditor{pending: make(map[uint32][]chan model.KillEvent)}
	ch, cancel := a.Expect(42)
	defer cancel()
	for _, event := range []model.KillEvent{
		{TargetPID: 42, Signal: 0, Origin: "sys_enter_kill"},
		{TargetPID: 42, Signal: 15, Origin: "sys_enter_kill"},
		{TargetPID: 42, Signal: 0, Origin: "signal_generate"},
		{TargetPID: 43, Signal: 15, Origin: "signal_generate"},
	} {
		a.dispatch(event)
	}
	select {
	case <-ch:
		t.Fatal("unrelated event confirmed a deletion")
	default:
	}
	a.dispatch(model.KillEvent{TargetPID: 42, Signal: 15, Origin: "signal_generate"})
	select {
	case event, ok := <-ch:
		if !ok || event.TargetPID != 42 {
			t.Fatal(event)
		}
	default:
		t.Fatal("missing confirmation")
	}
	if len(a.pending) != 0 {
		t.Fatal("pending entry leaked")
	}
}

func TestCancelRacesWithDispatch(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := &Auditor{pending: make(map[uint32][]chan model.KillEvent)}
		ch, cancel := a.Expect(42)
		var wg sync.WaitGroup
		wg.Add(2)
		go func() { defer wg.Done(); cancel(); cancel() }()
		go func() {
			defer wg.Done()
			a.dispatch(model.KillEvent{TargetPID: 42, Signal: 9, Origin: "signal_generate"})
		}()
		wg.Wait()
		for range ch {
		}
		if len(a.pending) != 0 {
			t.Fatal("pending entry leaked")
		}
	}
}

func TestCancelPreservesOtherWaiter(t *testing.T) {
	a := &Auditor{pending: make(map[uint32][]chan model.KillEvent)}
	first, cancel := a.Expect(42)
	second, cleanup := a.Expect(42)
	defer cleanup()
	cancel()
	if _, ok := <-first; ok {
		t.Fatal("cancelled channel open")
	}
	a.dispatch(model.KillEvent{TargetPID: 42, Signal: 9, Origin: "signal_generate"})
	if event, ok := <-second; !ok || event.Signal != 9 {
		t.Fatal("other waiter lost")
	}
}
