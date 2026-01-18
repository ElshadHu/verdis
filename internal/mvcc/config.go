package mvcc

import "regexp"

// RetentionPolicy defines how many versions to keep for keys matching a pattern
type RetentionPolicy struct {
	Pattern     *regexp.Regexp
	MaxVersions int
}

// Config holds MVCC engine configuration
type Config struct {
	// DefaultMaxVersions is the default number of versions to keep  per key
	DefaultMaxVersions int
	// RetentionPolicies allows per key pattern version limits
	RetentionPolicies []RetentionPolicy
	// TombstoneRetentionVersions is how many versions to keep tombstones
	TombstoneRetentionVersions int
	// EnableTimestampIndex enables version -> timestamp mapping
	EnableTimestampIndex bool
}

// DefaultConfig returns default configuration settings for development environment
func DefaultConfig() *Config {
	return &Config{
		DefaultMaxVersions:         1000,
		RetentionPolicies:          nil,
		TombstoneRetentionVersions: 100,
		EnableTimestampIndex:       true,
	}
}

// ProductionConfig returns recommended production settings
func ProductionConfig() *Config {
	return &Config{
		DefaultMaxVersions:         1000,
		TombstoneRetentionVersions: 1000,
		EnableTimestampIndex:       true,
		RetentionPolicies: []RetentionPolicy{
			{Pattern: regexp.MustCompile(`^audit:`), MaxVersions: 10000},
			{Pattern: regexp.MustCompile(`^cache:`), MaxVersions: 10},
			{Pattern: regexp.MustCompile(`^session:`), MaxVersions: 100},
		},
	}
}

// GetMaxVersionsForKey returns the max versions allowed for a specific key
func (c *Config) GetMaxVersionsForKey(key string) int {
	for _, policy := range c.RetentionPolicies {
		if policy.Pattern.MatchString(key) {
			return policy.MaxVersions
		}
	}
	return c.DefaultMaxVersions
}
