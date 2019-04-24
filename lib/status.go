package lib

import (
	"fmt"
	"strings"
)

type Status int

const (
	UNDEFINED Status = iota
	STARTED
	SUCCEEDED
	FAILED
)

func (s Status) Includes(wanted Status) bool {
	switch wanted {
	case STARTED:
		return STARTED <= s && s <= FAILED
	case SUCCEEDED:
		return s == SUCCEEDED
	case FAILED:
		return s == FAILED
	default:
		return false
	}
}

func (s Status) String() string {
	switch s {
	case UNDEFINED:
		return "undefined"
	case STARTED:
		return "started"
	case SUCCEEDED:
		return "succeeded"
	case FAILED:
		return "failed"
	default:
		return fmt.Sprintf("unknown-status-%d", s)
	}
}

func (s Status) MarshalText() ([]byte, error) {
	switch s {
	case UNDEFINED:
		return []byte("undefined"), nil
	case STARTED:
		return []byte("started"), nil
	case SUCCEEDED:
		return []byte("succeeded"), nil
	case FAILED:
		return []byte("failed"), nil
	default:
		return nil, fmt.Errorf("unknown status #%d", s)
	}
}

func (s *Status) UnmarshalText(text []byte) error {
	switch strings.TrimSpace(strings.ToLower(string(text))) {
	case "undefined":
		*s = UNDEFINED
	case "started":
		*s = STARTED
	case "succeeded":
		*s = SUCCEEDED
	case "failed":
		*s = FAILED
	default:
		return fmt.Errorf("unknown status: %s", text)
	}
	return nil
}
