package manager

import (
	"fmt"
	"math"
	"sync"
)

// alphabet defiens the character set used to shorten the URL (the longer
// alphabet gets, the shorter the URLs)
var alphabet []byte = []byte("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

var alphabetIndex []int64

func init() {
	alphabetIndex = make([]int64, 256, 256)
	for i := range alphabetIndex {
		alphabetIndex[i] = -1
	}
	for idx, c := range alphabet {
		alphabetIndex[c] = int64(idx)
	}
}

// MAXHASH defines the maximum number of shortened URLs the server can handle.
var maxHash int = 10000000

// From: https://golang.org/pkg/container/heap/#example__intHeap
// An IntHeap is a min-heap of ints.
type IntHeap []int32

// An IntHeap is a min-heap of ints.
// type IntHeap []int
func (h IntHeap) Len() int32           { return int32(len(h)) }
func (h IntHeap) Less(i, j int32) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int32)      { h[i], h[j] = h[j], h[i] }

func (h *IntHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length,
	// not just its contents.
	*h = append(*h, x.(int32))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// HashManager maintains the relation bewteen real URLs and hashes (shortened
// URLs). It is thread safe.
type HashManager struct {
	alias []string        // Maps the hashes to the URLs.
	urls  map[string]bool // All current URLs.
	holes IntHeap         // Maintains the minimum hash that is less than len(alias) and not used.
	lock  sync.RWMutex    // Allow concurrent reads and writes.
}

func encodeHash(hashInt int32) string {
	if hashInt == 0 {
		return "0"
	}
	base := int32(len(alphabet))

	var encode func() []byte
	encode = func() []byte {
		if hashInt == 0 {
			return []byte{}
		}
		remainder := hashInt % base
		hashInt /= base
		return append(encode(), alphabet[remainder])
	}

	return string(encode())
}

func decodeHash(hash string) int32 {
	base := int64(len(alphabet))
	var hashInt int64 = 0

	for _, c := range []byte(hash) {
		if idx := alphabetIndex[c]; idx != -1 {
			hashInt = hashInt*base + int64(idx)
			if hashInt > math.MaxInt32 {
				return -1
			}
		} else {
			return -1
		}
	}
	return int32(hashInt)
}

// NewHashManager returns an empty, unlocked HashManager.
func NewHashManager() *HashManager {
	return &HashManager{
		alias: []string{},
		urls:  make(map[string]bool),
		holes: IntHeap{},
	}
}

// GetFromInt returns the URL mapped on an integer.
// Error if hashInt is not known.
func (m *HashManager) get(hashInt int32) string {
	if hashInt >= int32(len(m.alias)) {
		return ""
	}
	return m.alias[hashInt]
}

// Get returns the URL associated with its hexadecimal hash.
// Error if hash can't be parsed as an hexstring or the hash is not known.
func (m *HashManager) Get(hash string) (url string, err error) {
	hashInt := decodeHash(hash)

	if hashInt == -1 {
		return "", fmt.Errorf("Hash %s in invalid", hash)
	}

	m.lock.RLock()
	defer m.lock.RUnlock()
	url = m.get(hashInt)
	if url == "" {
		err = fmt.Errorf("Can't find hash %s.", hash)
	}
	return
}

func (m *HashManager) setLowestFreeHash(url string) (hashInt int32) {
	if len(m.holes) != 0 {
		hashInt = m.holes.Pop().(int32)
		m.alias[hashInt] = url
		m.urls[url] = true
	} else if len(m.alias) <= maxHash {
		m.alias = append(m.alias, url)
		hashInt = int32(len(m.alias)) - 1
		m.urls[url] = true
	} else {
		hashInt = -1
	}
	return
}

// Add finds an available hash to be mapped to the given URL.
// Returns the found hash.
// Error if can't find a hash , or URL is already known.
func (m *HashManager) Add(url string) (hash string, err error) {
	if url == "" {
		return "", fmt.Errorf("Can't add the empty URL")
	}
	m.lock.Lock()
	defer m.lock.Unlock()

	if ok := m.urls[url]; ok {
		return "", fmt.Errorf("Can't add URL %s, already known", url)
	}

	hashInt := m.setLowestFreeHash(url)

	if hashInt == -1 {
		return "", fmt.Errorf(
			"Can't add URL %s, reached maximum capacity of %v URLs", url, maxHash,
		)
	}

	hash = encodeHash(hashInt)
	return
}

func (m *HashManager) delete(hashInt int32, url string) {
	m.alias[hashInt] = ""
	delete(m.urls, url)
	m.holes.Push(hashInt)
}

// Delete removes the relation from hash to its URL.
// Error if hash can't be parsed as an hexstring or it is unknown.
func (m *HashManager) Delete(hash string) (err error) {
	hashInt := decodeHash(hash)

	if hashInt == -1 {
		return fmt.Errorf("Hash %s invalid", hash)
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	url := m.get(hashInt)

	if url == "" {
		return fmt.Errorf("Can't delete hash %v, it's not known", hashInt)
	}

	m.delete(hashInt, url)
	return
}
