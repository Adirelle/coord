package lib

import (
	"log"
	"os"
	"sync"
	"testing"
	"time"
)

func TestLocalState_Get(t *testing.T) {
	l := log.New(os.Stdout, "TestLocalState_Get: ", 0)
	state := NewLocalState(l)
	defer state.Stop()

	if res := state.Get("/undefined"); res != UNDEFINED {
		t.Fatalf("Get on unset path should return UNDEFINED, not %#v", res)
	}
}

func TestLocalState_Put(t *testing.T) {
	l := log.New(os.Stdout, "TestLocalState_Put: ", 0)
	state := NewLocalState(l)
	defer func() { state.Stop() }()

	state.Put("/started", STARTED)

	if res := state.Get("/undefined"); res != UNDEFINED {
		t.Fatalf("Get on unset path should return UNDEFINED, not %#v", res)
	}

	if res := state.Get("/started"); res != STARTED {
		t.Fatalf("Get on '/started' should return STARTED, not %#v", res)
	}
}

func TestLocalState_Remove(t *testing.T) {
	l := log.New(os.Stdout, "TestLocalState_Remove: ", 0)
	state := NewLocalState(l)
	defer state.Stop()

	state.Put("/started", STARTED)

	if res := state.Get("/started"); res != STARTED {
		t.Fatalf("Get on '/started' should return STARTED, not %#v", res)
	}

	state.Remove("/started")

	if res := state.Get("/started"); res != UNDEFINED {
		t.Fatalf("Get on removed path should return UNDEFINED, not %#v", res)
	}
}

func TestLocalState_Wait(t *testing.T) {
	l := log.New(os.Stdout, "TestLocalState_Wait: ", 0)
	state := NewLocalState(l)
	defer state.Stop()

	state.Put("/started", STARTED)
	state.Put("/failed", FAILED)

	wg := sync.WaitGroup{}
	wg.Add(5)

	timeout := make(chan struct{})
	close(timeout)

	go func() {
		defer wg.Done()
		if !state.Wait("/started", predefinedStatusPredicates["started"], nil) {
			t.Errorf("Waiting for 'started' on '/started' should succeed")
		} else {
			t.Log("/started is started, as expected")
		}
	}()

	go func() {
		defer wg.Done()
		if !state.Wait("/deferred/start", predefinedStatusPredicates["started"], nil) {
			t.Errorf("Waiting for 'started' on '/deferred/start' should succeed")
		} else {
			t.Log("Waiting for 'started' on '/deferred/start' succeeded as expected")
		}
	}()

	go func() {
		defer wg.Done()
		if state.Wait("/failed", predefinedStatusPredicates["succeeded"], nil) {
			t.Errorf("Waiting for 'succeeded' on '/failed' should fail")
		} else {
			t.Log("Waiting for 'succeeded' on '/failed' failed as expected")
		}
	}()

	go func() {
		defer wg.Done()
		if state.Wait("/timeout", predefinedStatusPredicates["started"], timeout) {
			t.Errorf("Waiting for 'started' on '/timeout' should have failed on timeout")
		} else {
			t.Log("/timeout failed on time out, as expected")
		}
	}()

	go func() {
		defer wg.Done()
		time.Sleep(50 * time.Millisecond)
		state.Put("/deferred/start", STARTED)
	}()

	wg.Wait()
}
