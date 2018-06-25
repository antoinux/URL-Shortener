package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
)

// maxTry defines the maximum number of hash to try before giving up
// when adding a new URL.
var maxTry int = 1000

// HashManager maintains the relation bewteen real URLs and hashes (shortened
// URLs). It is thread safe.
type HashManager struct {
	alias map[uint64]string // Maps the hashes to the URLs.
	urls  map[string]bool   // All current URLs.
	lock  sync.RWMutex      // Allow concurrent reads and writes.
}

// NewHashManager returns an empty, unlocked HashManager.
func NewHashManager() HashManager {
	return HashManager{
		alias: make(map[uint64]string),
		urls:  make(map[string]bool),
	}
}

// GetFromInt returns the URL mapped on an integer.
// Error if hashInt is not known.
func (m *HashManager) GetFromInt(hashInt uint64) (url string, err error) {
	var ok bool
	m.lock.RLock()
	defer m.lock.RUnlock()
	m.getFromInt(hashInt)

	if !ok {
		err = fmt.Errorf("Integer Hash not found :%v.", hashInt)
		return
	}
	return
}

// Get returns the URL associated with its hexadecimal hash.
// Error if hash can't be parsed as an hexstring or the hash is not known.
func (m *HashManager) Get(hash string) (url string, err error) {
	var intHash uint64
	intHash, err = strconv.ParseUint(hash, 16, 64)

	if err != nil {
		return "", err
	}

	return m.GetFromInt(intHash)
}

// AddEntryFromInt adds the hashInt -> url mapping.
// Error if hashInt (or url) is already mapped.
func (m *HashManager) AddEntryFromInt(hashInt uint64, url string) (err error) {
	var ok bool
	var curUrl string
	m.lock.Lock()
	defer m.lock.Unlock()

	if curUrl, ok = m.alias[hashInt]; ok {
		err = fmt.Errorf("Integer hash already taken: mapped to %s.", curUrl)
		return
	}

	if m.urls[url] {
		err = fmt.Errorf("URL already taken.")
		return
	}

	m.alias[hashInt] = url
	m.urls[url] = true
	return
}

// AddEntry adds the hash -> url relation.
// Error if hash can't be parsed as an hexstring or hash (or url) is already taken.
func (m *HashManager) AddEntry(hash, url string) (err error) {
	var intHash uint64
	intHash, err = strconv.ParseUint(hash, 16, 64)

	if err != nil {
		return err
	}

	return m.AddEntryFromInt(intHash, url)
}

// Add finds an available hash to be mapped to the given URL.
// Returns the found hash.
// Error if can't find a hash after maxTry tries to find a hash, or
// URL is already known.
func (m *HashManager) Add(url string) (hash string, err error) {
	var ok bool
	m.lock.Lock()
	defer m.lock.Unlock()

	ok = m.urls[url]
	if ok {
		return "", fmt.Errorf("Can't add URL %s, already known", url)
	}

	var hashInt uint64 = 0
	_, ok = m.alias[hashInt]
	for tryCnt := 0; ok && tryCnt < maxTry; tryCnt++ {
		hashInt = rand.Uint64()
		_, ok = m.alias[hashInt]
	}

	if ok {
		return "", fmt.Errorf("Can't add URL %s after %v tries, the map may be too full", url, maxTry)
	}

	m.urls[url] = true
	m.alias[hashInt] = url
	hash = fmt.Sprintf("%X", hashInt)
	return
}

// DeleteFromInt removes the hashInt mapping to its URL.
// Error if hashInt is not mapped.
func (m *HashManager) DeleteFromInt(hashInt uint64) (err error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	url, ok := m.alias[hashInt]

	if !ok {
		err = fmt.Errorf("Can't delete hash %v, it's not known", hashInt)
		return
	}

	delete(m.alias, hashInt)
	delete(m.urls, url)
	return
}

// Delete removes the relation from hash to its URL.
// Error if hash can't be parsed as an hexstring or it is unknown.
func (m *HashManager) Delete(hash string) (err error) {
	var intHash uint64
	intHash, err = strconv.ParseUint(hash, 16, 64)

	if err != nil {
		return err
	}

	return m.DeleteFromInt(intHash)
}
