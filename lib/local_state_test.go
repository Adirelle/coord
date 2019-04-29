package lib

import (
	"errors"
	"log"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupState(t *testing.T) (*LocalState, func()) {
	l := log.New(os.Stdout, t.Name()+": ", 0)
	state := NewLocalState(l)
	return state, state.Stop
}

func TestLocalState_Get(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	assert.Equal(t, UNDEFINED, state.Get("/undefined"), "Get on unset path should return UNDEFINED")
}

func setter(s Status) StatusUpdateFunc {
	return StatusUpdateFunc(func(_ Status) (Status, error) {
		return s, nil
	})
}

func TestLocalState_Update(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	if !assert.Nil(t, state.Update("/started", setter(STARTED)), "unexpected error") {
		t.FailNow()
	}

	assert.Equal(t, STARTED, state.Get("/started"), "Get on '/started' should return STARTED")
}

func TestLocalState_Update_Failure(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	if assert.Error(
		t,
		state.Update("/started", StatusUpdateFunc(func(s Status) (Status, error) {
			return FAILED, errors.New("error")
		})),
	) {
		assert.Equal(t, UNDEFINED, state.Get("/started"), "Update should not change tree state on error")
	}

}

func TestLocalState_Update_Panic(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	assert.Error(
		t,
		state.Update("/started", StatusUpdateFunc(func(s Status) (Status, error) {
			panic("bla")
		})),
		"panic should be caught and converted to error"
	)
}

func TestLocalState_Remove(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	if err := state.Update("/started", setter(STARTED)); err != nil {
		t.Error("unexpected error", err)
	}
	if res := state.Get("/started"); res != STARTED {
		t.Fatalf("Get on '/started' should return STARTED, not %#v", res)
	}

	state.Remove("/started")

	if res := state.Get("/started"); res != UNDEFINED {
		t.Fatalf("Get on removed path should return UNDEFINED, not %#v", res)
	}
}

func TestLocalState_Wait(t *testing.T) {
	state, stop := setupState(t)
	defer stop()

	if err := state.Update("/started", setter(STARTED)); err != nil {
		t.Error("unexpected error", err)
	}
	if err := state.Update("/failed", setter(FAILED)); err != nil {
		t.Error("unexpected error", err)
	}

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
		if err := state.Update("/deferred/start", setter(STARTED)); err != nil {
			t.Error("unexpected error", err)
		}
	}()

	wg.Wait()
}
