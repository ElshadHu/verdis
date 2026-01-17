package mvcc

import "sync"

// Elshad: Yo BOI it is temp shi
// Dan: Some shi is goin' on here
// TODO: implement the MVCC storage interface with lock-free version chains
type Engine struct {
	data map[string][]byte // temporarily a simple map
	mu   sync.RWMutex
}

func NewEngine() *Engine {
	return &Engine{
		data: make(map[string][]byte),
	}
}

func (e *Engine) Get(key string) ([]byte, bool) {
	e.mu.RLock()
	val, ok := e.data[key]
	e.mu.RUnlock()
	return val, ok
}

func (e *Engine) Set(key string, value []byte) {
	e.mu.Lock()
	e.data[key] = value
	e.mu.Unlock()
}

func (e *Engine) Del(key string) bool {
	e.mu.Lock()
	_, existed := e.data[key]
	delete(e.data, key)
	e.mu.Unlock()
	return existed
}

func (e *Engine) Exists(key string) bool {
	e.mu.RLock()
	_, ok := e.data[key]
	e.mu.RUnlock()
	return ok
}
