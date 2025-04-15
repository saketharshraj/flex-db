package db

import (
	"errors"
	"sync"
	"time"
)

type ValueType int

const (
	TypeString ValueType = iota
	TypeList
	TypeHash
	// Future types can be added here
)

type Value struct {
	Type       ValueType
	Data       interface{}
	Expiration *time.Time // For TTL feature
}

// FlexDB is the main database structure
type FlexDB struct {
	data       map[string]Value
	lock       sync.RWMutex
	file       string
	writeQueue chan struct{}
}

// NewFlexDB initializes DB and loads data from disk
func NewFlexDB(filename string) *FlexDB {
	db := &FlexDB{
		data:       make(map[string]Value),
		file:       filename,
		writeQueue: make(chan struct{}, 100),
	}

	db.load()
	go db.writeLoop()
	go db.expirationChecker()
	return db
}

// expirationChecker periodically checks for expired keys
func (db *FlexDB) expirationChecker() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		keysToDelete := []string{}

		db.lock.RLock()
		for k, v := range db.data {
			if v.Expiration != nil && now.After(*v.Expiration) {
				keysToDelete = append(keysToDelete, k)
			}
		}
		db.lock.RUnlock()

		if len(keysToDelete) > 0 {
			db.lock.Lock()
			for _, k := range keysToDelete {
				delete(db.data, k)
			}
			db.lock.Unlock()
			db.triggerWrite()
		}
	}
}

// writeLoop handles periodic and triggered writes to disk
func (db *FlexDB) writeLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-db.writeQueue:
			select {
			case <-time.After(500 * time.Millisecond):
				db.save()
			case <-db.writeQueue:
				db.save()
			}
		case <-ticker.C:
			db.save()
		}
	}
}

// Set stores a string value with an optional expiration time
func (db *FlexDB) Set(key string, value string, expiration *time.Time) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.data[key] = Value{
		Type:       TypeString,
		Data:       value,
		Expiration: expiration,
	}
	db.triggerWrite()
}

// Get retrieves a value by key
func (db *FlexDB) Get(key string) (interface{}, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, ok := db.data[key]
	if !ok {
		return nil, errors.New("key not found")
	}

	// Check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		// Delete in a separate goroutine to avoid deadlock
		go func() {
			db.lock.Lock()
			delete(db.data, key)
			db.lock.Unlock()
			db.triggerWrite()
		}()
		return nil, errors.New("key not found")
	}

	return val.Data, nil
}

// Delete removes a key-value pair
func (db *FlexDB) Delete(key string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	if _, ok := db.data[key]; !ok {
		return errors.New("key not found")
	}
	delete(db.data, key)
	db.triggerWrite()
	return nil
}

// All returns a snapshot of all keys and values
func (db *FlexDB) All() map[string]interface{} {
	db.lock.RLock()
	defer db.lock.RUnlock()

	result := make(map[string]interface{})
	for k, v := range db.data {
		// Skip expired keys
		if v.Expiration != nil && time.Now().After(*v.Expiration) {
			continue
		}
		result[k] = v.Data
	}
	return result
}

// Expire sets an expiration time on a key
func (db *FlexDB) Expire(key string, duration time.Duration) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, ok := db.data[key]
	if !ok {
		return errors.New("key not found")
	}

	expiry := time.Now().Add(duration)
	val.Expiration = &expiry
	db.data[key] = val
	db.triggerWrite()
	return nil
}

// TTL returns the remaining time to live of a key with an expiration
func (db *FlexDB) TTL(key string) (time.Duration, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, ok := db.data[key]
	if !ok {
		return 0, errors.New("key not found")
	}

	if val.Expiration == nil {
		return -1, nil // Key exists but has no expiration
	}

	remaining := time.Until(*val.Expiration)

	return remaining, nil
}

func (db *FlexDB) Flush() {
	db.save()
}
