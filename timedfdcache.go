package timedfdcache

import (
	"os"
	"sync"
	"time"
)

type Cache struct {
	cache map[string]*CachedFile
	mut   sync.Mutex

	timeout time.Duration
}

// Creates a new cache which defers closing of any descriptors for the given
// period of time, allowing the descriptor to be reused within the given time
// frame.
func NewCache(duration time.Duration) *Cache {
	return &Cache{
		cache:   make(map[string]*CachedFile),
		timeout: duration,
	}
}

type CachedFile struct {
	*os.File
	parent *Cache
	timer  *time.Timer
}

// Schedule closure of the descriptor, which (subject it doesn't get reopened)
// will get closed after the timeout defined by the cache.
func (f *CachedFile) Close() error {
	f.parent.mut.Lock()
	defer f.parent.mut.Unlock()

	f.parent.cache[f.Name()] = f
	f.timer = time.AfterFunc(f.parent.timeout, func() {
		f.parent.mut.Lock()
		defer f.parent.mut.Unlock()

		delete(f.parent.cache, f.Name())
		f.File.Close()
	})
	return nil
}

// Open a new file descritpor, or reuse an old descriptor if available.
func (c *Cache) Open(path string) (*CachedFile, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	file, ok := c.cache[path]
	if ok && file.timer.Stop() {
		delete(c.cache, path)
		return file, nil
	}

	rfile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &CachedFile{
		File:   rfile,
		parent: c,
	}, nil
}
