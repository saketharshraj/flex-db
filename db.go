package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
	"time"
)

type FlexDB struct {
	data       map[string]string
	lock       sync.RWMutex
	file       string
	writeQueue chan struct{}
}

// NewFlexDB initializes DB and loads data from disk
func NewFlexDB(filename string) *FlexDB {
	db := &FlexDB{
		data:       make(map[string]string),
		file:       filename,
		writeQueue: make(chan struct{}, 100),
	}

	db.load()
	go db.writeLoop()
	return db
}

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

// load reads data from the file into memory
func (db *FlexDB) load() {
	db.lock.Lock()
	defer db.lock.Unlock()

	file, err := os.Open(db.file)
	if err != nil {
		if os.IsNotExist(err) {
			return // file doesn't exist, no problem
		}
		panic(err)
	}
	defer file.Close()

	bytes, _ := io.ReadAll(file)
	json.Unmarshal(bytes, &db.data)
}
func (db *FlexDB) triggerWrite() {
	select {
	case db.writeQueue <- struct{}{}:
		// successfully queued
	default:
		// queue is full — skip
	}
}

// save writes data to disk
func (db *FlexDB) save() {
	db.lock.RLock()
	defer db.lock.RUnlock()
	bytes, _ := json.MarshalIndent(db.data, "", "  ")
	_ = os.WriteFile(db.file, bytes, 0644)
}

// Set a key-value pair
func (db *FlexDB) Set(key, value string) {
	db.lock.Lock()
	defer db.lock.Unlock()

	db.data[key] = value
	db.triggerWrite()
}

// Get value by key
func (db *FlexDB) Get(key string) (string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, ok := db.data[key]
	if !ok {
		return "", errors.New("key not found")
	}
	return val, nil
}

// Delete a key-value pair
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

// All returns full DB snapshot
func (db *FlexDB) All() map[string]string {
	db.lock.RLock()
	defer db.lock.RUnlock()

	copy := make(map[string]string)
	for k, v := range db.data {
		copy[k] = v
	}
	return copy
}

func (db *FlexDB) Flush() {
	db.save()
}


