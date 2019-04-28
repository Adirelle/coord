package lib

import (
	"fmt"
	"testing"
)

var statuses = []Status{UNDEFINED, STARTED, SUCCEEDED, FAILED}

func TestPredefinedStatusUpdate_UnmarshalText(t *testing.T) {
	var u PredefinedStatusUpdate

	for n := range predefinedStatusUpdates {
		if err := u.UnmarshalText([]byte(n)); err != nil {
			t.Error("unexpected error:", err)
		}
	}

	if err := u.UnmarshalText([]byte("bla")); err == nil {
		t.Error("expected error")
	}
}

func ExampleStatusUpdate_Execute() {
	for _, s := range statuses {
		for n, a := range predefinedStatusUpdates {
			if r, e := a.Execute(s); e == nil {
				fmt.Printf("%s ==(%s)==> %s\n", s, n, r)
			}
		}
	}

	// Unordered output:
	// undefined ==(start)==> started
	// undefined ==(fail)==> failed
	// started ==(start)==> started
	// started ==(finish)==> succeeded
	// started ==(fail)==> failed
	// succeeded ==(finish)==> succeeded
	// failed ==(finish)==> failed
	// failed ==(fail)==> failed
}

func TestPredefinedStatusPredicate_UnmarshalText(t *testing.T) {
	var p PredefinedStatusPredicate

	for n := range predefinedStatusPredicates {
		if err := p.UnmarshalText([]byte(n)); err != nil {
			t.Error("unexpected error:", err)
		}
	}

	if err := p.UnmarshalText([]byte("bla")); err == nil {
		t.Error("expected error")
	}
}

func ExampleStatusPredicate() {
	for _, s := range statuses {
		for n, p := range predefinedStatusPredicates {
			if p.IsFulfilled(s) {
				fmt.Printf("%s is %s\n", s, n)
			} else if p.IsPossible(s) {
				fmt.Printf("%s can lead to %s\n", s, n)
			}
		}
	}

	// Unordered output:
	// undefined can lead to started
	// undefined can lead to running
	// undefined can lead to finished
	// undefined can lead to succeeded
	// undefined can lead to failed
	// started is started
	// started is running
	// started can lead to finished
	// started can lead to succeeded
	// started can lead to failed
	// succeeded is finished
	// succeeded is succeeded
	// failed is finished
	// failed is failed
}
