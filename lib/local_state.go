package lib

import (
	iradix "github.com/hashicorp/go-immutable-radix"
	"log"
)

type LocalState struct {
	root *iradix.Tree
	ctl  chan interface{}
	nc   chan struct{}
	l    *log.Logger
}

type putCommand struct {
	path   string
	status Status
}
type delCommand struct{ path string }

func NewLocalState(l *log.Logger) (s *LocalState) {
	s = &LocalState{
		iradix.New(),
		make(chan interface{}, 10),
		make(chan struct{}),
		l,
	}
	go s.loop()
	return s
}

func (s *LocalState) notify() {
	s.l.Println("notifying change")
	close(s.nc)
	s.nc = make(chan struct{})
}

func (s *LocalState) loop() {
	for input := range s.ctl {
		s.l.Printf("processing %#v", input)
		switch cmd := input.(type) {
		case putCommand:
			var prev interface{}
			s.root, prev, _ = s.root.Insert([]byte(cmd.path), cmd.status)
			if prev == nil || prev.(Status) != cmd.status {
				s.notify()
			}
		case delCommand:
			var prev interface{}
			s.root, prev, _ = s.root.Delete([]byte(cmd.path))
			if prev != nil {
				s.notify()
			}
		default:
			s.l.Printf("unknown LocalState command: %#v", input)
		}
	}
}

func (s *LocalState) Stop() {
	close(s.ctl)
}

func (s *LocalState) Put(path string, status Status) {
	s.ctl <- putCommand{path, status}
}

func (s *LocalState) Remove(path string) {
	s.ctl <- delCommand{path}
}

func (s *LocalState) Get(path string) Status {
	if data, found := s.root.Get([]byte(path)); found {
		return data.(Status)
	}
	return UNDEFINED
}

func (s *LocalState) Wait(path string, expected Status, timeout <-chan struct{}) bool {
	cond := func() bool {
		actual := s.Get(path)
		s.l.Printf("Expected=%s, actual=%s\n", expected, actual)
		return actual.Includes(expected)
	}
	for !cond() {
		s.l.Println("Waiting for change")
		select {
		case <-s.nc:
			s.l.Println("Been notified of a change")
		case <-timeout:
			s.l.Println("Timeout")
			return false
		}
	}
	return true
}
