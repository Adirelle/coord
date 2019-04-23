package main

import (
	"strings"
)

type Status int

const (
	UNDEFINED Status = iota
	STARTED
	SUCCEEDED
	FAILED
)

func (s Status) String() string {
	txt, err := s.MarshalText()
	if err != nil {
		return err.Error()
	}
	return string(txt)
}

func (s Status) MarshalText() ([]byte, error) {
	switch s {
	case STARTED:
		return []byte("started"), nil
	case SUCCEEDED:
		return []byte("succeeded"), nil
	case FAILED:
		return []byte("failed"), nil
	default:
		return []byte("undefined"), nil
	}
}

func (s *Status) UnmarshalText(text []byte) error {
	switch strings.TrimSpace(strings.ToLower(string(text))) {
	case "started":
		*s = STARTED
	case "succeeded":
		*s = SUCCEEDED
	case "failed":
		*s = FAILED
	default:
		*s = UNDEFINED
	}
	return nil
}
