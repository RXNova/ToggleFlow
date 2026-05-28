package api

import "sync"

// flagCache is a two-level in-memory store: projectID → environmentID → JSON bytes.
// Busted explicitly after any write that changes flag data for an environment.
type flagCache struct {
	mu    sync.RWMutex
	store map[int64]map[int64][]byte
}

func newFlagCache() *flagCache {
	return &flagCache{store: make(map[int64]map[int64][]byte)}
}

func (c *flagCache) get(projectID, envID int64) []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if envs, ok := c.store[projectID]; ok {
		return envs[envID]
	}
	return nil
}

func (c *flagCache) set(projectID, envID int64, data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.store[projectID] == nil {
		c.store[projectID] = make(map[int64][]byte)
	}
	c.store[projectID][envID] = data
}

// bust removes the cache for one environment — used when a single env changes
// (flag toggle, rules update).
func (c *flagCache) bust(projectID, envID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if envs, ok := c.store[projectID]; ok {
		delete(envs, envID)
	}
}

// bustProject removes all cached entries for a project — used when a change
// affects every environment (flag renamed, flag deleted, variations changed).
func (c *flagCache) bustProject(projectID int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.store, projectID)
}
