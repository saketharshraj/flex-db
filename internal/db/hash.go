package db

import (
	"errors"
	"fmt"
	"time"
)

// HSet sets the field in the hash stored at key to value.
// Returns 1 if the field is new, 0 if it was updated.
// Example: HSET user:1 name "John" -> 1
func (db *FlexDB) HSet(key, field, value string) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if exists {
		// Check if key has expired
		if val.Expiration != nil && time.Now().After(*val.Expiration) {
			delete(db.data, key)
			exists = false
		} else if val.Type != TypeHash {
			return 0, errors.New("value is not a hash")
		}
	}

	var hashMap map[string]string
	var fieldExists bool

	if exists {
		hashMap = val.Data.(map[string]string)
		_, fieldExists = hashMap[field]
	} else {
		hashMap = make(map[string]string)
		val = Value{
			Type: TypeHash,
			Data: hashMap,
		}
	}

	hashMap[field] = value
	val.Data = hashMap
	db.data[key] = val

	// Log to AOF if enabled
	if db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("HSET", key, field, value); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	if fieldExists {
		return 0, nil
	}
	return 1, nil
}

// HGet gets the value of a field in a hash.
// Returns an error if the key doesn't exist or is not a hash.
// Example: HGET user:1 name -> "John"
func (db *FlexDB) HGet(key, field string) (string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return "", errors.New("key not found")
	}

	if val.Type != TypeHash {
		return "", errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	value, exists := hashMap[field]
	if !exists {
		return "", errors.New("field not found")
	}

	return value, nil
}

// HDel removes fields from a hash.
// Returns the number of fields that were removed.
// Example: HDEL user:1 name age -> 2
func (db *FlexDB) HDel(key string, fields ...string) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return 0, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return 0, nil
	}

	if val.Type != TypeHash {
		return 0, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	deleted := 0

	for _, field := range fields {
		if _, exists := hashMap[field]; exists {
			delete(hashMap, field)
			deleted++
		}
	}

	if len(hashMap) == 0 {
		delete(db.data, key)
	} else {
		val.Data = hashMap
		db.data[key] = val
	}

	// Log to AOF if enabled and fields were deleted
	if deleted > 0 && db.aof != nil && db.aof.enabled {
		args := append([]string{key}, fields...)
		if err := db.aof.LogCommand("HDEL", args...); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	if deleted > 0 {
		db.triggerWrite()
	}

	return deleted, nil
}

// HGetAll returns all fields and values in a hash.
// Returns an empty map if the key doesn't exist.
// Example: HGETALL user:1 -> map[name:"John", age:"30"]
func (db *FlexDB) HGetAll(key string) (map[string]string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return map[string]string{}, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return map[string]string{}, nil
	}

	if val.Type != TypeHash {
		return nil, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	result := make(map[string]string, len(hashMap))
	for k, v := range hashMap {
		result[k] = v
	}

	return result, nil
}

// HExists checks if a field exists in a hash.
// Returns true if the field exists, false otherwise.
// Example: HEXISTS user:1 name -> true
func (db *FlexDB) HExists(key, field string) (bool, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return false, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return false, nil
	}

	if val.Type != TypeHash {
		return false, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	_, exists = hashMap[field]
	return exists, nil
}

// HLen returns the number of fields in a hash.
// Returns 0 if the key doesn't exist.
// Example: HLEN user:1 -> 2
func (db *FlexDB) HLen(key string) (int, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return 0, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return 0, nil
	}

	if val.Type != TypeHash {
		return 0, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	return len(hashMap), nil
}

// HKeys returns all fields in a hash.
// Returns an empty slice if the key doesn't exist.
// Example: HKEYS user:1 -> ["name", "age"]
func (db *FlexDB) HKeys(key string) ([]string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return []string{}, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return []string{}, nil
	}

	if val.Type != TypeHash {
		return nil, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	keys := make([]string, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}

	return keys, nil
}

// HVals returns all values in a hash.
// Returns an empty slice if the key doesn't exist.
// Example: HVALS user:1 -> ["John", "30"]
func (db *FlexDB) HVals(key string) ([]string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return []string{}, nil
	}

	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return []string{}, nil
	}

	if val.Type != TypeHash {
		return nil, errors.New("value is not a hash")
	}

	hashMap := val.Data.(map[string]string)
	values := make([]string, 0, len(hashMap))
	for _, v := range hashMap {
		values = append(values, v)
	}

	return values, nil
}
