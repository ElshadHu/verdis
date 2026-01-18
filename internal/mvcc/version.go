package mvcc

import (
	"sync"
	"sync/atomic"
	"time"
)

// VersionNode represents a single version of a key's value (immutable, lock-free as a result).
type VersionNode struct {
	// Version is monotonically increasing version ID
	Version uint64

	// Timestamp Unix nano timestamp when version was created
	Timestamp int64

	// Value is the actual data
	Value []byte

	// Deleted is a tombstone marker
	Deleted bool

	// Previous is a pointer to older version
	Prev *VersionNode
}

// VersionInfo is a read-only view of version metadata (for HISTORY command)
type VersionInfo struct {
	Version   uint64
	Timestamp int64
	Deleted   bool
	Size      int // len(Value)
}

// ToInfo creates a version info (read-only metadata) of the node
func (vn *VersionNode) ToInfo() VersionInfo {
	size := 0
	if vn.Value != nil {
		size = len(vn.Value)
	}
	return VersionInfo{
		Version:   vn.Version,
		Timestamp: vn.Timestamp,
		Deleted:   vn.Deleted,
		Size:      size,
	}
}

// Uses atomic operations  for lock-free version generation
type GlobalVersionManager struct {
	currentVersion atomic.Uint64
	// timestampMap stores version -> timestamp mapping
	timestampMap sync.Map
}

func NewGlobalVersionManager() *GlobalVersionManager {
	return &GlobalVersionManager{}
}

// Next version atomically increments and returns the next version number
func (gvm *GlobalVersionManager) NextVersion() (version uint64, timestamp int64) {
	version = gvm.currentVersion.Add(1)
	timestamp = time.Now().UnixNano()
	gvm.timestampMap.Store(version, timestamp)
	return version, timestamp
}

// GetTimestamp return the timestamp for the version
func (gvm *GlobalVersionManager) GetTimestamp(version uint64) (int64, bool) {
	if ts, ok := gvm.timestampMap.Load(version); ok {
		return ts.(int64), true
	}

	return 0, false
}

// PruneTimestamps removes timestamp entries older than the min version provided
func (gvm *GlobalVersionManager) PruneTimestamps(minVersion uint64) {
	gvm.timestampMap.Range(func(key, value any) bool {
		if key.(uint64) < minVersion {
			gvm.timestampMap.Delete(key)
		}
		return true
	})
}
