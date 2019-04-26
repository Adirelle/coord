package lib

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestLocalState_Get(t *testing.T) {
	l := log.New(os.Stdout,"test", 0)
	state := NewLocalState(l)
	defer state.Stop()

	if res := state.Get("/bla"); res != UNDEFINED {
		t.Fatalf("Get on unset path should return UNDEFINED, not %#v", res)
	}
}

func TestLocalState_Put(t *testing.T) {
	l := log.New(os.Stdout,"test", 0)
	state := NewLocalState(l)
	defer state.Stop()

	state.Put("/bla/thing", STARTED)

	time.Sleep(1 * time.Millisecond)

	if res := state.Get("/bla"); res != UNDEFINED {
		t.Fatalf("Get on unset path should return UNDEFINED, not %#v", res)
	}

	if res := state.Get("/bla/thing"); res != STARTED {
		t.Fatalf("Get on '/bla/thing' should return STARTED, not %#v", res)
	}
}

func TestLocalState_Remove(t *testing.T) {
	l := log.New(os.Stdout,"test", 0)
	state := NewLocalState(l)
	defer state.Stop()

	state.Put("/bla/thing", STARTED)

	time.Sleep(1 * time.Millisecond)

	if res := state.Get("/bla/thing"); res != STARTED {
		t.Fatalf("Get on '/bla/thing' should return STARTED, not %#v", res)
	}

	state.Remove("/bla/thing")

	time.Sleep(1 * time.Millisecond)

	if res := state.Get("/bla/thing"); res != UNDEFINED {
		t.Fatalf("Get on removed path should return UNDEFINED, not %#v", res)
	}
}

func TestLocalState_Wait(t *testing.T) {
	l := log.New(os.Stdout,"test: ", 0)
	state := NewLocalState(l)
	defer state.Stop()

	state.Put("/bla", STARTED)
	state.Put("/failed", FAILED)
	time.Sleep(10 * time.Millisecond)

	wg := sync.WaitGroup{}
	wg.Add(5)

	timeout := make(chan struct{})
	close(timeout)

	go func() {
		defer wg.Done()
		if !state.Wait("/bla", STARTED, nil) {
			t.Errorf("Wait should have succeeded on a path with expected status")
		}
	}()

	go func() {
		defer wg.Done()
		if !state.Wait("/bla/thing", STARTED, nil) {
			t.Errorf("Wait should have succeeded on a path when the status changes to the expected one")
		}
	}()

	go func() {
		defer wg.Done()
		if state.Wait("/failed", SUCCEEDED, nil) {
			t.Errorf("Wait should have succeeded on a path with incompatible status")
		}
	}()

	go func() {
		defer wg.Done()
		if state.Wait("/bla/thing", STARTED, timeout) {
			t.Errorf("Wait should have failed on timeout")
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		state.Put("/bla/thing", STARTED)
	}()

	wg.Wait()
}
