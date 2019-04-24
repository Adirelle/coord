package lib

import (
	"log"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type LocalState struct {
	root *iradix.Tree
	ctl  chan interface{}
	nc   chan interface{}
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
		make(chan interface{}),
		l,
	}
	go s.loop()
	return s
}

func (s *LocalState) notify() {
	s.l.Println("notifying change")
	close(s.nc)
	s.nc = make(chan interface{})
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

func (s *LocalState) Wait(path string, expected Status, timeout time.Duration) bool {
	t := time.After(timeout)
	for s.Get(path) != expected {
		s.l.Println("not the expected status, waiting for change")
		select {
		case <-s.nc:
			s.l.Println("data changed")
		case <-t:
			s.l.Println("timeout")
			return false
		}
	}
	s.l.Println("expected status !")
	return true
}
