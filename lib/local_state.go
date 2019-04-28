package lib

import (
	"log"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type LocalState struct {
	root    *iradix.Tree
	ctl     chan interface{}
	changed chan struct{}
	l       *log.Logger
}

type putCommand struct {
	path   string
	status Status
	done   chan struct{}
}

type delCommand struct {
	path string
	done chan struct{}
}

type stopCommand struct {
	done chan struct{}
}

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
	close(s.changed)
	s.changed = make(chan struct{})
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
			close(cmd.done)
		case delCommand:
			var prev interface{}
			s.root, prev, _ = s.root.Delete([]byte(cmd.path))
			if prev != nil {
				s.notify()
			}
			close(cmd.done)
		case stopCommand:
			close(s.ctl)
			close(cmd.done)
		default:
			s.l.Printf("unknown LocalState command: %#v", input)
		}
	}
}

func (s *LocalState) Stop() {
	done := make(chan struct{})
	s.ctl <- stopCommand{done}
	<-done
}

func (s *LocalState) Put(path string, status Status) {
	done := make(chan struct{})
	s.ctl <- putCommand{path, status, done}
	<-done
}

func (s *LocalState) Remove(path string) {
	done := make(chan struct{})
	s.ctl <- delCommand{path, done}
	<-done
}

func (s *LocalState) Get(path string) Status {
	if data, found := s.root.Get([]byte(path)); found {
		return data.(Status)
	}
	return UNDEFINED
}

func (s *LocalState) Wait(path string, predicate StatusPredicate, timeout <-chan struct{}) (ok bool) {
	shouldWait := func() bool {
		actual := s.Get(path)
		ok = predicate.IsFulfilled(actual)
		return !ok && predicate.IsPossible(actual)
	}
	for shouldWait() {
		select {
		case <-s.changed:
		case <-timeout:
			return false
		}
	}
	return
}
