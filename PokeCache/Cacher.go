package pokecache

import (
    "fmt"
    "time"
    "sync"
)

type CacheEntry struct{
    value []byte
    createdAt time.Time
}

type Cache struct{
    cache map[string]CacheEntry
    cacheEntryDeletionTime time.Duration
    mu sync.RWMutex
}
func NewCache(cacheEntryDeletionTimeSecs time.Duration) *Cache{
    c := Cache{make(map[string]CacheEntry),
                cacheEntryDeletionTimeSecs * time.Second,
                sync.RWMutex{}}
    go func (){
        for{ 
            time.Sleep(cacheEntryDeletionTimeSecs * time.Second)
            c.mu.Lock()
            for url,cache_entry := range c.cache{
                if cache_entry.createdAt.Add(cacheEntryDeletionTimeSecs * time.Second).Compare(time.Now()) == -1{
                    delete(c.cache,url)
                }
            }
            c.mu.Unlock()
        }
    }()
    return &c
}

func (c *Cache) Get(url string) ([]byte,error){
    c.mu.RLock()
    defer c.mu.RUnlock()
    val,ok := c.cache[url]
    if !ok{
        return nil,fmt.Errorf("%v not found in cache",url)
    }
    return val.value,nil
}

func (c *Cache) Add(url string, data []byte) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.cache[url] = CacheEntry{data,
                              time.Now()}
}

