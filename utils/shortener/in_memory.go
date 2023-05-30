package shortener

import (
	"sync"

	"github.com/srinkco/srink/utils/randomiser"
)

const HASH_NUM = 3

type InMemoryEngine struct {
	mu        sync.RWMutex
	hashToUrl map[string]string
	urlToHash map[string]string
}

func (e *InMemoryEngine) Shorten(url, hash string) string {
	if hash == "" {
		if hash, ok := e.urlToHash[url]; ok {
			return hash
		}
		hash = randomiser.GetString(HASH_NUM)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.hashToUrl[hash] = url
	e.urlToHash[url] = hash
	return hash
}

func (e *InMemoryEngine) GetUrl(hash string) (url string) {
	e.mu.RLock()
	url = e.hashToUrl[hash]
	e.mu.RUnlock()
	return
}
