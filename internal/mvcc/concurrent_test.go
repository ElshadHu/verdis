package mvcc_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/ElshadHu/verdis/internal/mvcc"
)

// --- Test Configuration ---

type casTestConfig struct {
	numWriters int
	numKeys    int // 0 = single key test
}

type rwTestConfig struct {
	numWriters     int
	numReaders     int
	readsPerReader int
}

type gapTestConfig struct {
	numWriters int
	numKeys    int
}

var (
	// Fast tests: always run, CI-friendly
	fastSingleKey   = casTestConfig{numWriters: 1_000, numKeys: 0}
	fastMultiKey    = casTestConfig{numWriters: 1_000, numKeys: 100}
	fastReadWrite   = rwTestConfig{numWriters: 50, numReaders: 50, readsPerReader: 100}
	fastVersionGaps = gapTestConfig{numWriters: 1_000, numKeys: 10}
	fastIntegrity   = casTestConfig{numWriters: 1_000, numKeys: 0}
)

// --- Fast Tests (always run) ---

func TestCASContention_SingleKey(t *testing.T) {
	runSingleKeyContention(t, fastSingleKey)
}

func TestCASContention_MultiKey(t *testing.T) {
	runMultiKeyContention(t, fastMultiKey)
}

func TestCASContention_WritersAndReaders(t *testing.T) {
	runWritersAndReaders(t, fastReadWrite)
}

func TestCASContention_NoVersionGaps(t *testing.T) {
	runNoVersionGaps(t, fastVersionGaps)
}

func TestCASContention_ChainIntegrity(t *testing.T) {
	runChainIntegrity(t, fastIntegrity)
}

// --- Core Test Logic ---

// runSingleKeyContention verifies N concurrent writers to the same key all succeed.
func runSingleKeyContention(t *testing.T, cfg casTestConfig) {
	t.Helper()
	engine := mvcc.NewEngine()
	const key = "contention-key"

	versions := make([]uint64, cfg.numWriters)
	var wg sync.WaitGroup
	wg.Add(cfg.numWriters)

	for i := range cfg.numWriters {
		go func(idx int) {
			defer wg.Done()
			value := fmt.Appendf(nil, "value-%d", idx)
			versions[idx] = engine.Set(key, value)
		}(i)
	}
	wg.Wait()

	chainVersions := collectVersionChain(t, engine, key)
	assertChainComplete(t, versions, chainVersions)
}

// runMultiKeyContention verifies contention across N distinct keys.
func runMultiKeyContention(t *testing.T, cfg casTestConfig) {
	t.Helper()
	engine := mvcc.NewEngine()

	versionsPerKey := make(map[string][]uint64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(cfg.numWriters)

	for i := range cfg.numWriters {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", idx%cfg.numKeys)
			value := fmt.Appendf(nil, "value-%d", idx)
			version := engine.Set(key, value)

			mu.Lock()
			versionsPerKey[key] = append(versionsPerKey[key], version)
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	for keyIdx := range cfg.numKeys {
		key := fmt.Sprintf("key-%d", keyIdx)
		chainVersions := collectVersionChain(t, engine, key)
		assertChainComplete(t, versionsPerKey[key], chainVersions)
	}
}

// runWritersAndReaders verifies concurrent readers always see consistent state.
func runWritersAndReaders(t *testing.T, cfg rwTestConfig) {
	t.Helper()
	engine := mvcc.NewEngine()
	const key = "rw-contention-key"

	// seed with initial value so readers don't hit empty chain
	engine.Set(key, []byte("initial"))

	var wg sync.WaitGroup
	wg.Add(cfg.numWriters + cfg.numReaders)

	var consistencyErrors atomic.Int32

	for i := range cfg.numWriters {
		go func(idx int) {
			defer wg.Done()
			value := fmt.Appendf(nil, "writer-%d-value", idx)
			engine.Set(key, value)
		}(i)
	}

	for i := range cfg.numReaders {
		go func(idx int) {
			defer wg.Done()
			for range cfg.readsPerReader {
				value, ok := engine.Get(key)
				if ok {
					if value == nil {
						consistencyErrors.Add(1)
						t.Errorf("reader %d: ok=true but nil value", idx)
						return
					}
					if len(value) == 0 {
						consistencyErrors.Add(1)
						t.Errorf("reader %d: got empty value", idx)
						return
					}
				}
			}
		}(i)
	}

	wg.Wait()

	chainVersions := collectVersionChain(t, engine, key)
	expectedLen := 1 + cfg.numWriters // initial + writers
	if len(chainVersions) != expectedLen {
		t.Errorf("chain length: expected %d, got %d", expectedLen, len(chainVersions))
	}

	versionSet := make(map[uint64]bool)
	for _, v := range chainVersions {
		if versionSet[v] {
			t.Errorf("duplicate version %d in chain", v)
		}
		versionSet[v] = true
	}

	if errors := consistencyErrors.Load(); errors > 0 {
		t.Errorf("detected %d consistency errors during concurrent reads", errors)
	}
}

// runNoVersionGaps verifies no gaps in global version sequence.
func runNoVersionGaps(t *testing.T, cfg gapTestConfig) {
	t.Helper()
	engine := mvcc.NewEngine()

	var wg sync.WaitGroup
	wg.Add(cfg.numWriters)

	allVersions := make(chan uint64, cfg.numWriters)

	for i := range cfg.numWriters {
		go func(idx int) {
			defer wg.Done()
			key := fmt.Sprintf("gap-test-key-%d", idx%cfg.numKeys)
			value := fmt.Appendf(nil, "value-%d", idx)
			version := engine.Set(key, value)
			allVersions <- version
		}(i)
	}

	wg.Wait()
	close(allVersions)

	versionSet := make(map[uint64]bool)
	var maxVersion uint64
	for v := range allVersions {
		versionSet[v] = true
		if v > maxVersion {
			maxVersion = v
		}
	}

	if len(versionSet) != int(maxVersion) {
		t.Errorf("expected %d unique versions, got %d", maxVersion, len(versionSet))
	}

	for v := uint64(1); v <= maxVersion; v++ {
		if !versionSet[v] {
			t.Errorf("missing version %d in global sequence", v)
		}
	}
}

// runChainIntegrity verifies chain is fully traversable with contiguous versions.
func runChainIntegrity(t *testing.T, cfg casTestConfig) {
	t.Helper()
	engine := mvcc.NewEngine()
	const key = "integrity-key"

	var wg sync.WaitGroup
	wg.Add(cfg.numWriters)

	for i := range cfg.numWriters {
		go func(idx int) {
			defer wg.Done()
			value := fmt.Appendf(nil, "value-%d", idx)
			engine.Set(key, value)
		}(i)
	}

	wg.Wait()

	history, err := engine.History(key, 0)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != cfg.numWriters {
		t.Errorf("expected %d nodes, got %d", cfg.numWriters, len(history))
	}

	seen := make(map[uint64]bool)
	for _, info := range history {
		if seen[info.Version] {
			t.Errorf("duplicate version %d", info.Version)
		}
		seen[info.Version] = true
	}

	for v := uint64(1); v <= uint64(cfg.numWriters); v++ {
		if !seen[v] {
			t.Errorf("version %d missing from chain", v)
		}
	}
}

// --- Helpers ---

// collectVersionChain walks the version chain, returning all version numbers.
func collectVersionChain(t *testing.T, engine *mvcc.Engine, key string) []uint64 {
	t.Helper()
	history, err := engine.History(key, 0)
	if err != nil {
		t.Fatalf("failed to get history for key %s: %v", key, err)
	}

	versions := make([]uint64, len(history))
	for i, info := range history {
		versions[i] = info.Version
	}
	return versions
}

// assertChainComplete verifies all written versions are in the chain with no duplicates.
func assertChainComplete(t *testing.T, written, chain []uint64) {
	t.Helper()

	if len(chain) != len(written) {
		t.Errorf("chain length mismatch: expected %d, got %d", len(written), len(chain))
	}

	writtenSet := make(map[uint64]bool, len(written))
	for _, v := range written {
		writtenSet[v] = true
	}

	chainSet := make(map[uint64]bool, len(chain))
	for _, v := range chain {
		if chainSet[v] {
			t.Errorf("duplicate version %d in chain", v)
		}
		chainSet[v] = true
	}

	for v := range writtenSet {
		if !chainSet[v] {
			t.Errorf("version %d written but missing from chain", v)
		}
	}

	for v := range chainSet {
		if !writtenSet[v] {
			t.Errorf("phantom version %d in chain (never written)", v)
		}
	}
}
