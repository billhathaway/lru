// lru_test
package lru

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

var keys []string

func init() {
	keys = make([]string, 5000000)
	for i := 0; i < 5000000; i++ {
		keys[i] = strconv.Itoa(i)
	}
}

// Test_simpleFoundCase verifies that a key,value set will be returned correctly
func Test_simpleFoundCase(t *testing.T) {
	l, err := NewLru(10)
	assert.Nil(t, err)
	l.Set("a", "b")
	val, found := l.Get("a")
	assert.True(t, found)
	assert.Equal(t, val, "b")
}

// Test_simpleNotFoundCase verifies that the bool returned is false when the key isn't found
func Test_simpleNotFoundCase(t *testing.T) {
	l, err := NewLru(10)
	assert.Nil(t, err)
	_, found := l.Get("a")
	assert.False(t, found)
}

// Test_simpleExpireCase verifies that with a limit of N entries, the first entry will be expired once N additional entries are added
func Test_simpleExpireCase(t *testing.T) {
	size := 10
	l, err := NewLru(uint(size))
	assert.Nil(t, err)

	l.Set("willExpire", "test")
	val, found := l.Get("willExpire")
	assert.True(t, found)
	assert.Equal(t, val, "test")

	for i := 0; i < size; i++ {
		stringNum := strconv.Itoa(i)
		l.Set(stringNum, stringNum)
		val, found = l.Get(stringNum)
		assert.True(t, found)
		assert.Equal(t, val, stringNum)
	}

	val, found = l.Get("willExpire")
	assert.False(t, found)
	assert.Nil(t, val)
}

// Test_hitRate verifies that hitrate is calculated correctly
func Test_hitRate(t *testing.T) {
	l, err := NewLru(10)
	assert.Nil(t, err)

	assert.Equal(t, 0.0, l.HitRate())

	l.Set("a", "a")
	l.Get("a")
	assert.Equal(t, 1.0, l.HitRate())

	l.Get("b")
	assert.Equal(t, 0.5, l.HitRate())
}

func Test_update(t *testing.T) {
	l, err := NewLru(10)
	assert.Nil(t, err)

	prev := l.Set("a", "initialValue")
	assert.Nil(t, prev)

	prev = l.Set("a", "newValue")
	assert.Equal(t, prev, "initialValue")
	assert.Equal(t, l.Len(), 1)

	l.Set("b", "newValue")
	assert.Equal(t, l.Len(), 2)
}

func Benchmark_insertExpire(b *testing.B) {
	l, _ := NewLru(10)
	for i := 0; i < b.N; i++ {
		l.Set(keys[i%len(keys)], 100)
	}
}

func Benchmark_insertNoExpire(b *testing.B) {
	l, _ := NewLru(10000000)
	for i := 0; i < b.N; i++ {
		l.Set(keys[i%len(keys)], 100)
	}
}
func Benchmark_GetFound(b *testing.B) {
	l, _ := NewLru(uint(b.N))
	for i := 0; i < b.N; i++ {
		l.Set(keys[i%len(keys)], 100)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.Get(keys[i%len(keys)])
	}
}
func Benchmark_GetNotFound(b *testing.B) {
	l, _ := NewLru(uint(b.N))
	for i := 0; i < b.N; i++ {
		l.Set(keys[i%len(keys)], 100)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stringNum := strconv.Itoa(b.N + i)
		l.Get(stringNum)
	}
}

func lruReader(count int, l *lru, wg *sync.WaitGroup) {
	for i := 0; i < count; i++ {
		l.Get(keys[i])
	}
	wg.Done()
}

func Benchmark_getMultiGoRoutines(b *testing.B) {
	l, _ := NewLru(uint(b.N))
	for i := 0; i < b.N; i++ {
		l.Set(keys[i], 100)
	}
	wg := &sync.WaitGroup{}
	b.ResetTimer()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go lruReader(b.N, l, wg)
	}
	wg.Wait()
}
