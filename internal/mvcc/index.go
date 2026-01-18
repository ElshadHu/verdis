package mvcc

import (
	"sync"
	"sync/atomic"
)

// VersionChainHead wraps atomic pointer to the head of a version chain.
type VersionChainHead struct {
	head atomic.Pointer[VersionNode]
}

// Load returns the current head of the version chain
func (vch *VersionChainHead) Load() *VersionNode {
	return vch.head.Load()
}

// CompareAndSwap automically updates head if it matches expected
func (vch *VersionChainHead) CompareAndSwap(expected, new *VersionNode) bool {
	return vch.head.CompareAndSwap(expected, new)
}

// Index is a lock-free map from keys to version chain heads.
type Index struct {
	data sync.Map
}

func NewIndex() *Index {
	return &Index{}
}

// GetOrCreateChain gets the version chain head for a key (creates it if key doesn't exist)
func (idx *Index) GetOrCreateChain(key string) *VersionChainHead {
	if existing, ok := idx.data.Load(key); ok {
		return existing.(*VersionChainHead)
	}

	newChain := &VersionChainHead{}
	actual, _ := idx.data.LoadOrStore(key, newChain)
	return actual.(*VersionChainHead)
}

// GetChain returns the version chain head for a key (nil if doesn't exist)
func (idx *Index) GetChain(key string) *VersionChainHead {
	if existing, ok := idx.data.Load(key); ok {
		return existing.(*VersionChainHead)
	}
	return nil
}

// Keys returns all keys in the index
func (idx *Index) Keys() []string {
	var keys []string
	idx.data.Range(func(key, value any) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

// Count  returns approximate number of keys
func (idx *Index) Count() int {
	count := 0
	idx.data.Range(func(key, value any) bool {
		count++
		return true
	})
	return count
}
