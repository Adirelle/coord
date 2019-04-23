package main

import (
	"sync"
	"time"

	iradix "github.com/hashicorp/go-immutable-radix"
)

type LocalState struct {
	tree *iradix.Tree
	mu   sync.RWMutex
	cond *sync.Cond
}

func NewLocalState() (s *LocalState) {
	s = &LocalState{}
	s.tree = iradix.New()
	s.cond = sync.NewCond(&s.mu)
	return
}

func (s *LocalState) Get(path string) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	if data, found := s.tree.Get([]byte(path)); found {
		return data.(Status)
	}
	return UNDEFINED
}

func (s *LocalState) Put(path string, status Status) <-chan bool {
	c := make(chan bool)
	go func() {
		s.mu.Lock()
		var changed bool
		s.tree, _, changed = s.tree.Insert([]byte(path), status)
		if changed {
			s.cond.Broadcast()
		}
		s.mu.Unlock()
		c <- changed
		close(c)
	}()
	return c
}

func (s *LocalState) Remove(path string) <-chan bool {
	c := make(chan bool)
	go func() {
		s.mu.Lock()
		var changed bool
		s.tree, _, changed = s.tree.Delete([]byte(path))
		if changed {
			s.cond.Broadcast()
		}
		s.mu.Unlock()
		c <- changed
		close(c)
	}()
	return c
}

func (s *LocalState) Wait(path string, expected Status, timeout time.Duration) bool {
	waiting := true
	c := make(chan bool)
	time.AfterFunc(timeout, func() {
		waiting = false
		c <- false
	})
	go func() {
		s.cond.L.Lock()
		defer s.cond.L.Unlock()
		defer close(c)
		for waiting {
			if data, found := s.tree.Get([]byte(path)); found && data.(Status) == expected {
				c <- true
				return
			}
			s.cond.Wait()
		}
	}()
	return <-c
}
