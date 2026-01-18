package mvcc

import (
	verr "github.com/ElshadHu/verdis/internal/errors"
)

var (
	ErrKeyNotFound     = verr.ErrKeyNotFound
	ErrVersionNotFound = verr.ErrVersionNotFound
	ErrKeyDeleted      = verr.ErrKeyDeleted
)

type Engine struct {
	index          *Index
	versionManager *GlobalVersionManager
	config         *Config
}

// NewEngine creates a new MVCC engine with DEFAULT config
func NewEngine() *Engine {
	return NewEngineWithConfig(DefaultConfig())
}

// NewEngineWithConfig creates a new MVCC engine with given config
func NewEngineWithConfig(config *Config) *Engine {
	return &Engine{
		index:          NewIndex(),
		versionManager: NewGlobalVersionManager(),
		config:         config,
	}
}

// Get returns the latest value for a key
func (e *Engine) Get(key string) ([]byte, bool) {
	chain := e.index.GetChain(key)
	if chain == nil {
		return nil, false
	}
	head := chain.Load()
	if head == nil {
		return nil, false
	}

	// if latest version is tombstone key is deleted
	if head.Deleted {
		return nil, false
	}

	return head.Value, true
}

// Set stores a value for a key, creating a new version
func (e *Engine) Set(key string, value []byte) uint64 {
	version, timestamp := e.versionManager.NextVersion()

	newNode := &VersionNode{
		Version:   version,
		Timestamp: timestamp,
		Value:     value,
		Deleted:   false,
		Prev:      nil,
	}

	chain := e.index.GetOrCreateChain(key)

	// CAS loop to try until prepend successful
	for {
		currentHead := chain.Load()
		newNode.Prev = currentHead

		if chain.CompareAndSwap(currentHead, newNode) {
			// prepended current version
			break
		}

		// another writer won, retry with their version as the Prev
	}

	return version
}

// Del marks a key as deleted by adding a tombstone version
func (e *Engine) Del(key string) bool {
	chain := e.index.GetChain(key)
	if chain == nil {
		return false
	}

	currentHead := chain.Load()
	if currentHead == nil {
		return false // key has no versions
	}

	// if already deleted, still creates new tombstone
	version, timestamp := e.versionManager.NextVersion()

	tombstone := &VersionNode{
		Version:   version,
		Timestamp: timestamp,
		Value:     nil,
		Deleted:   true,
		Prev:      nil,
	}

	// CAS loop to prepend tombstone
	for {
		currentHead := chain.Load()
		tombstone.Prev = currentHead

		if chain.CompareAndSwap(currentHead, tombstone) {
			break
		}
	}

	return true
}

// Exists checks if a key exists and is not deleted
func (e *Engine) Exists(key string) bool {
	chain := e.index.GetChain(key)
	if chain == nil {
		return false
	}
	head := chain.Load()

	if head == nil {
		return false
	}
	return !head.Deleted
}

// GetAtVersion returns the value at a specific version or earlier
func (e *Engine) GetAtVersion(key string, version uint64) ([]byte, error) {
	chain := e.index.GetChain(key)
	if chain == nil {
		return nil, ErrKeyNotFound
	}

	head := chain.Load()
	if head == nil {
		return nil, ErrKeyNotFound
	}

	// walk chain backwards until we find the version <= requested
	current := head
	for current != nil {
		if current.Version <= version {
			if current.Deleted {
				return nil, ErrKeyDeleted
			}
			return current.Value, nil
		}
		current = current.Prev
	}

	return nil, ErrVersionNotFound
}

// History returns version meta a key
func (e *Engine) History(key string, maxVersions int) ([]VersionInfo, error) {
	chain := e.index.GetChain(key)
	if chain == nil {
		return nil, ErrKeyNotFound
	}

	head := chain.Load()
	if head == nil {
		return nil, ErrKeyNotFound
	}

	var history []VersionInfo
	current := head

	for current != nil {
		history = append(history, current.ToInfo())
		if maxVersions > 0 && len(history) > maxVersions {
			break
		}
		current = current.Prev
	}
	return history, nil
}

// CurrentVersion returns the global version counter (for snapshots)
func (e *Engine) CurrentVersion() uint64 {
	return e.versionManager.CurrentVersion()
}

// EngineStats holds engine statistics
type EngineStats struct {
	KeyCount       int
	CurrentVersion uint64
}

// Stats returns engine statistics (for INFO command)
func (e *Engine) Stats() EngineStats {
	return EngineStats{
		KeyCount:       e.index.Count(),
		CurrentVersion: e.versionManager.CurrentVersion(),
	}
}
