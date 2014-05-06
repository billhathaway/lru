// Cache project Cache.go
package lru

import (
	"container/list"
	"errors"
	"log"
	"sync"
)

type Cache struct {
	data    map[string]*list.Element
	list    *list.List
	limit   uint
	mutex   sync.Mutex
	hits    uint
	misses  uint
	expired uint
	removes uint
	logger  *log.Logger
	purger  Purger
}

type Stats struct {
	Hits    uint
	Misses  uint
	Limit   uint
	Len     uint
	Expired uint
	Removes uint
}

type cacheEntry struct {
	key   string
	value interface{}
}

// Optional interface for cached objects
type Purger interface {
	OnPurge(key string, value interface{})
}

// New creates a new Cache cache
func New(limit uint) (*Cache, error) {
	if limit == 0 {
		return nil, errors.New("limit must be positive")
	}
	Cache := new(Cache)
	Cache.data = make(map[string]*list.Element)
	Cache.list = list.New()
	Cache.limit = limit
	return Cache, nil
}

func (l *Cache) RegisterPurger(purger Purger) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.purger = purger
}

func (l *Cache) RemoveAll() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for l.list.Len() > 0 {
		l.expire()
	}
}

// expire removes the oldest entry.  The mutex lock is already help by Set.
func (l *Cache) expire() {
	entry := l.list.Back()
	if entry != nil {
		l.expired++
		ce := entry.Value.(cacheEntry)
		delete(l.data, ce.key)
		l.list.Remove(entry)
		if l.purger != nil {
			l.purger.OnPurge(ce.key, ce.value)
		}
	} else {
		// shouldn't be here unless something else is wrong
		l.logger.Printf("Cache - nil entry when trying to remove, limit=%d len=%d\n", l.limit, l.list.Len())
	}
}

// Set adds the value and sets it to the head of the list.
// If the key was already present, the entry is updated and the previous value is returned.
// If the key was not already present, nil is returned
func (l *Cache) Set(key string, val interface{}) interface{} {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	for l.list.Len() >= int(l.limit) {
		l.expire()
	}
	if entry, found := l.data[key]; found {
		l.list.MoveToFront(entry)
		ce := entry.Value.(cacheEntry)
		previousValue := ce.value
		ce.value = val
		//fmt.Printf("updating entry for %s\n", key)
		return previousValue
	}
	ce := cacheEntry{key, val}
	entry := l.list.PushFront(ce)
	l.data[key] = entry
	return nil
}

// Get returns the value if it exists and true, otherwise returns nil and false
// the entry is moved to the front of the list if it is found
func (l *Cache) Get(key string) (interface{}, bool) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if entry, found := l.data[key]; found {
		l.hits++
		l.list.MoveToFront(entry)
		l.data[key] = entry
		return entry.Value.(cacheEntry).value, true
	}
	l.misses++
	return nil, false
}

// Remove removes a key, returning true if the key was found, false if it was not
func (l *Cache) Remove(key string) bool {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.removes++
	if entry, found := l.data[key]; found {
		l.list.Remove(entry)
		delete(l.data, key)
		return true
	}
	return false
}

// SetLogger sets the logger.  There is currently only a single log statement in the package.
func (l *Cache) SetLogger(logger *log.Logger) {
	l.logger = logger
}

// Stats returns a stats structure containing information on the cache hits, misses, max size, current size, expired entries, and entries removed
func (l *Cache) Stats() Stats {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return Stats{l.hits, l.misses, l.limit, uint(l.list.Len()), l.expired, l.removes}
}

// ResetStats resets the hit,misses, and expired counters
func (l *Cache) ResetStats() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.hits = 0
	l.misses = 0
	l.expired = 0
	l.removes = 0

}

// Limit returns the maximum number of entries that may be kept in the cache
func (l *Cache) Limit() uint {
	return l.limit
}

// Len returns the number of entries in the cache
func (l *Cache) Len() int {
	return l.list.Len()
}

// HitRate returns a number between 0.0 and 1.0 indicating the percentage of get calls that were found in the cache
func (l *Cache) HitRate() float32 {
	if l.hits == 0 {
		return 0.0
	}
	return float32(l.hits) / float32(l.hits+l.misses)
}
