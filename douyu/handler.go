package douyu

import (
	"fmt"
	"sync"
)

// Handler: message function wrapper
type Handler func(*Message)

// HandlerWrapper: message handler wrapper
type HandlerWrapper struct {
	handle  Handler
	enabled bool
	name    string
}

// Run: message handler callback
func (s *HandlerWrapper) Run(msg *Message) {
	if s.enabled {
		s.handle(msg)
	}
}

func (s *HandlerWrapper) getName() string {
	return s.name
}

func (s *HandlerWrapper) enableHandle() {
	s.enabled = true
	return
}

func (s *HandlerWrapper) disableHandle() {
	s.enabled = false
	return
}

// HandlerRegister: message handler manager
type HandlerRegister struct {
	mu   sync.RWMutex
	hmap map[string][]*HandlerWrapper
}

// CreateHandlerRegister: create handler register
func CreateHandlerRegister() *HandlerRegister {
	return &HandlerRegister{
		hmap: make(map[string][]*HandlerWrapper),
	}
}

// Add: add message callback handle to handler register
func (hr *HandlerRegister) Add(key string, h Handler, name string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	for _, v := range hr.hmap {
		for _, handle := range v {
			if handle.getName() == name {
				return fmt.Errorf("handler name %s has been registered", name)
			}
		}
	}
	hr.hmap[key] = append(hr.hmap[key], &HandlerWrapper{handle: h, enabled: true, name: name})
	return nil
}

// Get: get message handler
func (hr *HandlerRegister) Get(key string) (error, []*HandlerWrapper) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	if v, ok := hr.hmap[key]; ok {
		return nil, v
	}
	return fmt.Errorf("no handlers for key [%s]", key), nil
}

// EnableByType: enable handler by message type
func (hr *HandlerRegister) EnableByType(key string) error {
	err, handles := hr.Get(key)
	if err != nil {
		return err
	}
	hr.mu.Lock()
	defer hr.mu.Unlock()
	// all
	for _, v := range handles {
		v.enableHandle()
	}
	return nil
}

// DisableByType: disable handler by message type
func (hr *HandlerRegister) DisableByType(key string) error {
	err, handles := hr.Get(key)
	if err != nil {
		return err
	}
	hr.mu.Lock()
	defer hr.mu.Unlock()
	// all
	for _, v := range handles {
		v.disableHandle()
	}
	return nil
}

// EnableByName: enable message handler by name
func (hr *HandlerRegister) EnableByName(name string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	for _, handles := range hr.hmap {
		for _, v := range handles {
			if v.getName() == name {
				v.enableHandle()
				return nil
			}
		}
	}
	return fmt.Errorf("cannot find handler %s", name)
}

// DisableByName: disable message handler by name
func (hr *HandlerRegister) DisableByName(name string) error {
	hr.mu.Lock()
	defer hr.mu.Unlock()
	for _, handles := range hr.hmap {
		for _, v := range handles {
			if v.getName() == name {
				v.disableHandle()
				return nil
			}
		}
	}
	return fmt.Errorf("cannot find handler %s", name)
}

// Dump: output all message handlers
func (hr *HandlerRegister) Dump() string {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	str := "[plugins dump]\n"
	for k, handles := range hr.hmap {
		for _, v := range handles {
			str += fmt.Sprintf("%d %s [%v]\n", k, v.getName(), v.enabled)
		}
	}
	return str
}
