// lru project lru.go
package lru

import (
	"container/list"
	"errors"
	"log"
	"sync"
)

type lru struct {
	data    map[string]*list.Element
	list    *list.List
	limit   uint
	mutex   sync.Mutex
	hits    uint
	misses  uint
	expired uint
	logger  *log.Logger
}

type Stats struct {
	Hits    uint
	Misses  uint
	Limit   uint
	Len     uint
	Expired uint
}

type cacheEntry struct {
	key   string
	value interface{}
}

// New creates a new LRU cache
func New(limit uint) (*lru, error) {
	if limit == 0 {
		return nil, errors.New("limit must be positive")
	}
	lru := new(lru)
	lru.data = make(map[string]*list.Element)
	lru.list = list.New()
	lru.limit = limit
	return lru, nil
}

// expire removes the oldest entry.  The mutex lock is already help by Set.
func (l *lru) expire() {
	entry := l.list.Back()
	if entry != nil {
		l.expired++
		ce := entry.Value.(cacheEntry)
		delete(l.data, ce.key)
		l.list.Remove(entry)
	} else {
		// shouldn't be here unless something else is wrong
		l.logger.Printf("lru - nil entry when trying to remove, limit=%d len=%d\n", l.limit, l.list.Len())
	}
}

// Set adds the value and sets it to the head of the list.
// If the key was already present, the entry is updated and the previous value is returned.
// If the key was not already present, nil is returned
func (l *lru) Set(key string, val interface{}) interface{} {
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
func (l *lru) Get(key string) (interface{}, bool) {
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

// SetLogger sets the logger.  There is currently only a single log statement in the package.
func (l *lru) SetLogger(logger *log.Logger) {
	l.logger = logger
}

// Stats returns a stats structure containing information on the cache hits, misses, max size, current size, and expired entries
func (l *lru) Stats() Stats {
	return Stats{l.hits, l.misses, l.limit, uint(l.list.Len()), l.expired}
}

// ResetStats resets the hit,misses, and expired counters
func (l *lru) ResetStats() {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.hits = 0
	l.misses = 0
	l.expired = 0

}

// Limit returns the maximum number of entries that may be kept in the cache
func (l *lru) Limit() uint {
	return l.limit
}

// Len returns the number of entries in the cache
func (l *lru) Len() int {
	return l.list.Len()
}

// HitRate returns a number between 0.0 and 1.0 indicating the percentage of get calls that were found in the cache
func (l *lru) HitRate() float32 {
	if l.hits == 0 {
		return 0.0
	}
	return float32(l.hits) / float32(l.hits+l.misses)
}
