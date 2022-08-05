package signals

import (
	"reflect"
	"sync"
)

type ISubscribeable interface {
	On(name string, f interface{})
	Off(name string)
}

type IEmitable interface {
	Emit(name string, v ...interface{})
}
type Signal struct {
	rwMutex  sync.RWMutex
	handlers map[string][]reflect.Value
}

func New() *Signal {
	return &Signal{
		handlers: make(map[string][]reflect.Value),
	}
}

func (s *Signal) On(name string, f interface{}) {
	s.rwMutex.Lock()
	if _, ok := s.handlers[name]; !ok {
		s.handlers[name] = make([]reflect.Value, 0, 1)
	}

	r := reflect.ValueOf(f)
	s.handlers[name] = append(s.handlers[name], r)
	s.rwMutex.Unlock()
}
func (s *Signal) Off(name string) {
	s.rwMutex.Lock()
	delete(s.handlers, name)
	s.rwMutex.Unlock()
}

func (s *Signal) Emit(name string, v ...interface{}) {
	s.rwMutex.RLock()
	go func() {
		defer s.rwMutex.RUnlock()

		params := make([]reflect.Value, len(v))

		for i, value := range v {
			params[i] = reflect.ValueOf(value)
		}

		if handlers, ok := s.handlers[name]; ok {

			for _, handler := range handlers {
				go handler.Call(params)
			}
		}
	}()
}
