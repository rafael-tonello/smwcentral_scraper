package misc

import (
	"sync"
)

type CallbackInfo[T any] struct {
	f  func(data T)
	id int
}

type Stream[T any] struct {
	listeners map[int](*CallbackInfo[T])
	lastValue T
	idCount   int

	lock sync.Mutex
}

func NewStream[T any]() *Stream[T] {
	ret := Stream[T]{idCount: 0, listeners: map[int](*CallbackInfo[T]){}}

	return &ret
}

func NewStreamWithInitialValue[T any](initialValue T) *Stream[T] {
	ret := Stream[T]{lastValue: initialValue, idCount: 0}
	return &ret
}

func (s *Stream[T]) Listen(f func(data T)) int {
	s.lock.Lock()
	s.listeners[s.idCount] = &(CallbackInfo[T]{f: f, id: s.idCount})
	s.lock.Unlock()
	s.idCount += 1
	return s.idCount - 1
}

func (s *Stream[T]) StopListen(observerId int) {
	s.lock.Lock()
	_, found := s.listeners[observerId]

	if found {
		delete(s.listeners, observerId)
	}
	s.lock.Unlock()
}

func (s *Stream[T]) Stream(data T) {
	s.lastValue = data
	s.lock.Lock()
	for _, f := range s.listeners {
		tmp := f.f

		go tmp(data)
	}
	s.lock.Unlock()
}

func (s *Stream[T]) GetLast() T {
	return s.lastValue
}

// helper functions
func (s *Stream[T]) Subscribe(f func(data T)) int { return s.Listen(f) }

func (s *Stream[T]) Unsubscribe(observerId int) { s.StopListen(observerId) }

func (s *Stream[T]) Publish(data T) { s.Stream(data) }
func (s *Stream[T]) Add(data T)     { s.Stream(data) }
