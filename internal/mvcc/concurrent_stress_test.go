//go:build stress

package mvcc_test

import "testing"

// Stress test configs (1M writers)
var (
	stressSingleKey   = casTestConfig{numWriters: 1_000_000, numKeys: 0}
	stressMultiKey    = casTestConfig{numWriters: 1_000_000, numKeys: 10_000}
	stressReadWrite   = rwTestConfig{numWriters: 100_000, numReaders: 1_000, readsPerReader: 100}
	stressVersionGaps = gapTestConfig{numWriters: 1_000_000, numKeys: 100}
	stressIntegrity   = casTestConfig{numWriters: 1_000_000, numKeys: 0}
)

// --- Stress Tests (only with -tags stress) ---

func TestCASContention_SingleKey_Stress(t *testing.T) {
	runSingleKeyContention(t, stressSingleKey)
}

func TestCASContention_MultiKey_Stress(t *testing.T) {
	runMultiKeyContention(t, stressMultiKey)
}

func TestCASContention_WritersAndReaders_Stress(t *testing.T) {
	runWritersAndReaders(t, stressReadWrite)
}

func TestCASContention_NoVersionGaps_Stress(t *testing.T) {
	runNoVersionGaps(t, stressVersionGaps)
}

func TestCASContention_ChainIntegrity_Stress(t *testing.T) {
	runChainIntegrity(t, stressIntegrity)
}
