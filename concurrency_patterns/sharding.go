package concurrency_patterns

import (
	"crypto/sha1"
	"fmt"
	"sync"
)

// Struct with a map and a
// composed sync.RWMutex
var items = struct {
	sync.RWMutex
	m map[string]int
}{m: make(map[string]int)}

func ThreadSafeRead(key string) int {
	items.RLock()
	value := items.m[key]
	items.RUnlock()
	return value
}
func ThreadSafeWrite(key string, value int) {
	items.Lock()
	items.m[key] = value
	items.Unlock()
}

// Vertical sharding

type Shard struct {
	sync.RWMutex                        // Compose from sync.RWMutex
	m            map[string]interface{} // m contains the shard's data

}

type ShardedMap []*Shard // ShardedMap is a *Shards slice

func NewShardedMap(nshards int) ShardedMap {
	shards := make([]*Shard, nshards) // Initialize a *Shards slice
	for i := 0; i < nshards; i++ {
		shard := make(map[string]interface{})
		shards[i] = &Shard{m: shard}
	}
	return shards // A ShardedMap IS a *Shards slice! }
}

func (m ShardedMap) getShardIndex(key string) int {
	checksum := sha1.Sum([]byte(key)) // Use Sum from "crypto/sha1"

	// this will only create 255 shards if you want more use
	// hash := int(sum[13]) << 8 | int(sum[17])
	hash := int(checksum[17]) // Pick an arbitrary byte as the hash
	return hash % len(m)      // Mod by len(m) to get index
}

func (m ShardedMap) getShard(key string) *Shard {
	index := m.getShardIndex(key)
	return m[index]
}

func (m ShardedMap) Get(key string) interface{} {
	shard := m.getShard(key)
	shard.RLock()
	defer shard.RUnlock()
	return shard.m[key]
}

func (m ShardedMap) Set(key string, value interface{}) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()
	shard.m[key] = value
}

func (m ShardedMap) Delete(key string) {
	shard := m.getShard(key)
	shard.Lock()
	defer shard.Unlock()
	delete(shard.m, key)
}

func (m ShardedMap) Contains(key string) bool {
	shard := m.getShard(key)
	shard.RLock()
	defer shard.RUnlock()
	_, ok := shard.m[key]
	return ok
}

// to establish lock on all keys, maybe to get all the keys that are present in all shards, we'll get it concurrently using waitgroup

func (m ShardedMap) Keys() []string {
	keys := make([]string, 0) // Create an empty keys slice

	mutex := sync.Mutex{} // Mutex for write safety to keys

	wg := sync.WaitGroup{}
	wg.Add(len(m)) // Create a wait group and add a wait value for each slice of Shard
	for _, shard := range m {
		go func(s *Shard) { // Run a goroutine for each slice
			s.RLock() // Establish a read lock on s

			for key := range s.m { // Get the slice's keys
				mutex.Lock()
				keys = append(keys, key)
				mutex.Unlock()
			}
			s.RUnlock() // Release the read lock
			wg.Done()   // Tell the WaitGroup it's done
		}(shard)
	}
	wg.Wait()   // Block until all reads are done return keys
	return keys // Return combined keys slice
}

func TestVerticalSharding() {
	shardedMap := NewShardedMap(5)
	shardedMap.Set("alpha", 1)
	shardedMap.Set("beta", 2)
	shardedMap.Set("gamma", 3)
	fmt.Println(shardedMap.Get("alpha"))
	fmt.Println(shardedMap.Get("beta"))
	fmt.Println(shardedMap.Get("gamma"))
	keys := shardedMap.Keys()
	for _, k := range keys {
		fmt.Println(k)
	}
}
