Implements a LRU cache for string keys and interface{} values
==

Example:
--
```
import "github.com/billhathaway/lru"
import "fmt" 

cache := lru.New(1000)
cache.put("key1","value1")
cache.put("key2",789)

key := "key3"
value,found := cache.get(key)

if found {
	fmt.Printf("key %s found, value=%v\n",key,value)
} else {
	fmt.Printf("key %s not found\n",key)
}
```
