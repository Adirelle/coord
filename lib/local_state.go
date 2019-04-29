package lib

import (
	"errors"
	"fmt"
	"log"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type LocalState struct {
	root    *iradix.Tree
	ctl     chan interface{}
	changed chan struct{}
	l       *log.Logger
}

type updateCommand struct {
	path   string
	update StatusUpdate
	done   chan error
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
		case updateCommand:
			err := s.update([]byte(cmd.path), cmd.update)
			go func() {
				cmd.done <- err
			}()
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

func errorify(data interface{}) error {
	if data == nil {
		return nil
	}
	switch value := data.(type) {
	case error:
		return value
	case fmt.Stringer:
		return errors.New(value.String())
	case string:
		return errors.New(value)
	default:
		return fmt.Errorf("%#v", value)
	}
}

func safeCall(f func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			err = errorify(rec)
		}
	}()
	err = f()
	return
}

func (s *LocalState) update(path []byte, update StatusUpdate) (err error) {
	status := UNDEFINED
	if data, found := s.root.Get(path); found {
		status = data.(Status)
	}
	var newStatus Status
	err = safeCall(func() (err error) {
		newStatus, err = update.Execute(status)
		return
	})
	if err == nil && status != newStatus {
		s.root, _, _ = s.root.Insert(path, newStatus)
		s.notify()
	}
	return
}

func (s *LocalState) Stop() {
	done := make(chan struct{})
	s.ctl <- stopCommand{done}
	<-done
}

func (s *LocalState) Update(path string, update StatusUpdate) error {
	done := make(chan error)
	s.ctl <- updateCommand{path, update, done}
	return <-done
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
