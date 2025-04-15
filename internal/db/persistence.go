package db

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// PersistentValue is used for serialization
type PersistentValue struct {
	Type       ValueType   `json:"type"`
	Data       interface{} `json:"data"`
	Expiration int64       `json:"exp,omitempty"` // Unix timestamp
}

// load reads data from the file into memory
func (db *FlexDB) load() {
	db.lock.Lock()
	defer db.lock.Unlock()

	file, err := os.Open(db.file)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		return
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return
	}

	// Temporary map for deserialization
	tempData := make(map[string]PersistentValue)
	if err := json.Unmarshal(bytes, &tempData); err != nil {
		return
	}

	// Convert to runtime format
	now := time.Now()
	for k, v := range tempData {
		var exp *time.Time
		if v.Expiration > 0 {
			t := time.Unix(v.Expiration, 0)
			exp = &t
			// Skip expired keys
			if now.After(t) {
				continue
			}
		}
		
		db.data[k] = Value{
			Type:       v.Type,
			Data:       v.Data,
			Expiration: exp,
		}
	}
}

// save writes data to disk
func (db *FlexDB) save() {
	db.lock.RLock()
	defer db.lock.RUnlock()

	// Convert to serializable format
	tempData := make(map[string]PersistentValue)
	for k, v := range db.data {
		pv := PersistentValue{
			Type: v.Type,
			Data: v.Data,
		}
		
		if v.Expiration != nil {
			pv.Expiration = v.Expiration.Unix()
		}
		
		tempData[k] = pv
	}

	bytes, err := json.MarshalIndent(tempData, "", "  ")
	if err != nil {
		return
	}
	
	// Use atomic file write to prevent corruption
	tempFile := db.file + ".tmp"
	if err := os.WriteFile(tempFile, bytes, 0644); err != nil {
		return
	}
	os.Rename(tempFile, db.file)
}

func (db *FlexDB) triggerWrite() {
	select {
	case db.writeQueue <- struct{}{}:
		// successfully queued
	default:
		// queue is full — skip
	}
}