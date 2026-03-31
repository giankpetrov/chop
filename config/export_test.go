package config

import "sync"

// ResetCacheForTest zeros all sync.Once path and config caches.
// Call this at the start of any test that manipulates env vars
// (XDG_CONFIG_HOME, HOME, etc.) to prevent stale cached results.
func ResetCacheForTest() {
	configDirOnce = sync.Once{}
	configDirVal = ""
	dataDirOnce = sync.Once{}
	dataDirVal = ""
	globalCfgOnce = sync.Once{}
	globalCfg = Config{}
	globalFiltersOnce = sync.Once{}
	globalFilters = nil
}
