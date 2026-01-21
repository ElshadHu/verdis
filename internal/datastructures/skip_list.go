package datastructures

import (
	"bytes"
	"sync/atomic"
	"time"
)

const (
	// Supports millions
	maxLevel    = 16
	probability = 4
)

// SkipListNode represents a node in the lock-free skip list
type SkipListNode struct {
	key     []byte
	value   atomic.Pointer[[]byte]
	forward []atomic.Pointer[SkipListNode]
	marked  atomic.Bool // logical deletion flag
	level   int
}

func newSkipListNode(key, value []byte, level int) *SkipListNode {
	node := &SkipListNode{
		key:     key,
		forward: make([]atomic.Pointer[SkipListNode], level+1),
		level:   level,
	}

	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	node.value.Store(&valueCopy)
	return node
}

// SkipList is a lock-free concurrent sorted key-value store
type SkipList struct {
	head  *SkipListNode
	level atomic.Int32
	size  atomic.Int64
	seed  atomic.Uint64
}

func NewSkipList() *SkipList {
	head := &SkipListNode{
		forward: make([]atomic.Pointer[SkipListNode], maxLevel),
		level:   maxLevel - 1,
	}
	sl := &SkipList{
		head: head,
	}

	sl.seed.Store(uint64(time.Now().UnixNano()))
	return sl
}

// randomLevel generates a random level using lock-free xorshift64
func (sl *SkipList) randomLevel() int {
	level := 0
	for {
		old := sl.seed.Load()
		// xorshift64 algorithm
		next := old ^ (old << 13)
		next ^= next >> 7
		next ^= next << 17
		if sl.seed.CompareAndSwap(old, next) {
			// Use lower bits to find the level
			// Each level has 1/4 probability
			r := next
			for level < maxLevel-1 && (r&3) == 0 {
				level++
				r >>= 2
			}
			break
		}
		// CAS failed, another goroutine updated seed, retry it
	}
	return level
}

// findPath finds predecessors and successors at each level for a key
func (sl *SkipList) findPath(key []byte) ([maxLevel]*SkipListNode, [maxLevel]*SkipListNode) {
	var preds, succs [maxLevel]*SkipListNode
	current := sl.head
	for i := maxLevel - 1; i >= 0; i-- {
		for {
			next := current.forward[i].Load()
			if next == nil {
				break
			}

			if next.marked.Load() {
				nextNext := next.forward[i].Load()
				current.forward[i].CompareAndSwap(next, nextNext)
				continue
			}
			if bytes.Compare(next.key, key) >= 0 {
				break
			}
			current = next
		}
		preds[i] = current
		succs[i] = current.forward[i].Load()
	}

	return preds, succs
}

// Get retrieves a value by key
func (sl *SkipList) Get(key []byte) ([]byte, bool) {
	current := sl.head
	for i := maxLevel - 1; i >= 0; i-- {
		for {
			next := current.forward[i].Load()
			if next == nil {
				break
			}

			// Help unlink marked nodes
			if next.marked.Load() {
				nextNext := next.forward[i].Load()
				current.forward[i].CompareAndSwap(next, nextNext)
				continue
			}

			cmp := bytes.Compare(next.key, key)
			if cmp > 0 {
				break
			}
			if cmp == 0 {
				if next.marked.Load() {
					return nil, false
				}
				val := next.value.Load()
				if val == nil {
					return nil, false
				}

				// Prevent mutation by returning a copy
				result := make([]byte, len(*val))
				copy(result, *val)
				return result, true
			}
			current = next
		}
	}

	return nil, false
}

// Put inserts or updates a key-value pair
func (sl *SkipList) Put(key, value []byte) bool {
	// Make copies to avoid external mutation
	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	for {
		preds, succs := sl.findPath(keyCopy)

		// Check if key already exists at level 0
		if succs[0] != nil && bytes.Equal(succs[0].key, keyCopy) && !succs[0].marked.Load() {
			// Update existing value atomically
			valueCopy := make([]byte, len(value))
			copy(valueCopy, value)
			succs[0].value.Store(&valueCopy)
			return false
		}

		// Generate random level for new node
		newLevel := sl.randomLevel()

		// Update skip list's max level if needed
		for {
			currLevel := sl.level.Load()
			if int32(newLevel) <= currLevel {
				break
			}
			if sl.level.CompareAndSwap(currLevel, int32(newLevel)) {
				break
			}
		}

		// Ensure preds for higher levels point to head
		for i := int(sl.level.Load()); i >= 0; i-- {
			if preds[i] == nil {
				preds[i] = sl.head
			}
		}

		// Create new node
		newNode := newSkipListNode(keyCopy, value, newLevel)

		// Set forward pointers before linking
		for i := 0; i <= newLevel; i++ {
			newNode.forward[i].Store(succs[i])
		}

		// Link at level 0 first - this is the linearization point
		if !preds[0].forward[0].CompareAndSwap(succs[0], newNode) {
			// Another writer modified, retry from scratch
			continue
		}

		// Successfully linked at level 0, now link upper levels
		for i := 1; i <= newLevel; i++ {
			for {
				pred := preds[i]
				succ := succs[i]

				// Refresh forward pointer in case structure changed
				newNode.forward[i].Store(succ)

				if pred.forward[i].CompareAndSwap(succ, newNode) {
					break
				}

				// Structure changed, refind predecessors
				preds, succs = sl.findPath(keyCopy)

				// Check if our node was deleted by another thread
				if succs[0] == nil || !bytes.Equal(succs[0].key, keyCopy) || succs[0].marked.Load() {
					// Node was deleted, upper links don't matter
					sl.size.Add(1)
					return true
				}
			}
		}

		sl.size.Add(1)
		return true
	}
}

// Delete removes a key from the skip list
func (sl *SkipList) Delete(key []byte) bool {
	preds, succs := sl.findPath(key)
	target := succs[0]

	// key not found
	if target == nil || !bytes.Equal(target.key, key) {
		return false
	}

	// already deleted
	if target.marked.Load() {
		return false
	}

	// logical deletion: atomically mark the node
	if !target.marked.CompareAndSwap(false, true) {
		return false
	}

	// physical deletion: unlink from all levels
	for i := target.level; i >= 0; i-- {
		for {
			pred := preds[i]
			if pred == nil {
				pred = sl.head
			}

			succ := pred.forward[i].Load()
			if succ != target {
				break
			}

			targetNext := target.forward[i].Load()
			if pred.forward[i].CompareAndSwap(target, targetNext) {
				break
			}

			preds, _ = sl.findPath(key)
		}
	}

	sl.size.Add(-1)
	return true
}

// Contains checks if key exists in list
func (sl *SkipList) Contains(key []byte) bool {
	_, found := sl.Get(key)
	return found
}

// Size returns the number of entries in the skip list.
func (sl *SkipList) Size() int64 {
	return sl.size.Load()
}

type SkipListIterator struct {
	sl      *SkipList
	current *SkipListNode
}

// NewIterator creates a new iterator positioned before the first element.
func (sl *SkipList) NewIterator() *SkipListIterator {
	return &SkipListIterator{
		sl:      sl,
		current: sl.head,
	}
}

// Next advances the iterator to the next non-deleted entry.
func (it *SkipListIterator) Next() bool {
	for {
		next := it.current.forward[0].Load()
		if next == nil {
			return false
		}
		it.current = next

		// Skip logically deleted nodes
		if !next.marked.Load() {
			return true
		}
	}
}

// Key returns the current entry's key.
func (it *SkipListIterator) Key() []byte {
	if it.current == nil || it.current.key == nil {
		return nil
	}
	// Return copy to prevent mutation
	result := make([]byte, len(it.current.key))
	copy(result, it.current.key)
	return result
}

// Value returns the current entry's value.
func (it *SkipListIterator) Value() []byte {
	if it.current == nil {
		return nil
	}
	val := it.current.value.Load()
	if val == nil {
		return nil
	}
	// Return copy to prevent mutation
	result := make([]byte, len(*val))
	copy(result, *val)
	return result
}

// Valid returns true if the iterator is at a valid position.
func (it *SkipListIterator) Valid() bool {
	return it.current != nil &&
		it.current != it.sl.head &&
		it.current.key != nil &&
		!it.current.marked.Load()
}

// SeekToFirst creates an iterator positioned at the first entry.
func (sl *SkipList) SeekToFirst() *SkipListIterator {
	it := sl.NewIterator()
	it.Next()
	return it
}

// Seek positions the iterator at the first key >= target.
func (sl *SkipList) Seek(target []byte) *SkipListIterator {
	it := &SkipListIterator{
		sl:      sl,
		current: sl.head,
	}

	current := sl.head
	for i := maxLevel - 1; i >= 0; i-- {
		for {
			next := current.forward[i].Load()
			if next == nil {
				break
			}
			// Help unlink marked nodes
			if next.marked.Load() {
				nextNext := next.forward[i].Load()
				current.forward[i].CompareAndSwap(next, nextNext)
				continue
			}
			if bytes.Compare(next.key, target) >= 0 {
				break
			}
			current = next
		}
	}

	it.current = current

	for it.Next() {
		if bytes.Compare(it.current.key, target) >= 0 {
			return it
		}
	}

	return it
}

// Range iterates over all keys in [start, end).
func (sl *SkipList) Range(start, end []byte, fn func(key, value []byte) bool) {
	it := sl.Seek(start)

	for it.Valid() {
		key := it.Key()

		if end != nil && bytes.Compare(key, end) >= 0 {
			break
		}

		value := it.Value()
		if !fn(key, value) {
			break
		}

		if !it.Next() {
			break
		}
	}
}

// Clear removes all entries from the skip list.
func (sl *SkipList) Clear() {
	// Mark all existing nodes as deleted first
	current := sl.head.forward[0].Load()
	for current != nil {
		current.marked.Store(true)
		current = current.forward[0].Load()
	}
	for i := range maxLevel {
		sl.head.forward[i].Store(nil)
	}
	sl.level.Store(0)
	sl.size.Store(0)
}
