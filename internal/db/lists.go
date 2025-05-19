package db

import (
	"errors"
	"fmt"
	"time"
)

// LPush inserts values at the beginning of a list
func (db *FlexDB) LPush(key string, values ...string) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// check if key exists in db
	val, exists := db.data[key]

	if exists {
		// check if key has expired
		if val.Expiration != nil && time.Now().After(*val.Expiration) {
			delete(db.data, key)
			exists = false
		} else if val.Type != TypeList {
			return 0, errors.New("value is not a list")
		}
	}

	var list []string

	// get existing list or create a new one
	if exists {
		list = val.Data.([]string)
	} else {
		list = make([]string, 0, len(values))
		val = Value{
			Type: TypeList,
			Data: list,
		}
	}

	// prepend values in reverse order
	for i := len(values) - 1; i >= 0; i-- {
		list = append([]string{values[i]}, list...)
	}

	val.Data = list
	db.data[key] = val

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		args := append([]string{key}, values...)
		if err := db.aof.LogCommand("LPUSH", args...); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return len(list), nil
}

// RPush appends values to the end of a list
func (db *FlexDB) RPush(key string, values ...string) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	// check if key exists in db
	val, exists := db.data[key]

	if exists {
		// check if key has expired
		if val.Expiration != nil && time.Now().After(*val.Expiration) {
			delete(db.data, key)
			exists = false
		} else if val.Type != TypeList {
			return 0, errors.New("value is not a list")
		}
	}

	var list []string

	// get existing list or create a new one
	if exists {
		list = val.Data.([]string)
	} else {
		list = make([]string, 0, len(values))
		val = Value{
			Type: TypeList,
			Data: list,
		}
	}

	// append values
	list = append(list, values...)

	val.Data = list
	db.data[key] = val

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		args := append([]string{key}, values...)
		if err := db.aof.LogCommand("RPUSH", args...); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return len(list), nil
}

// LPop removes and returns the first element of a list
func (db *FlexDB) LPop(key string) (string, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		delete(db.data, key)
		return "", errors.New("key not found")
	}

	if val.Type != TypeList {
		return "", errors.New("value is not a list")
	}

	list := val.Data.([]string)
	if len(list) == 0 {
		return "", errors.New("list is empty")
	}

	// get first element and remove it
	item := list[0]
	list = list[1:]

	// if list is empty after pop, delete the key
	if len(list) == 0 {
		delete(db.data, key)
	} else {
		val.Data = list
		db.data[key] = val
	}

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("LPOP", key); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return item, nil
}

// RPop removes and returns the last element of a list
func (db *FlexDB) RPop(key string) (string, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		delete(db.data, key)
		return "", errors.New("key not found")
	}

	if val.Type != TypeList {
		return "", errors.New("value is not a list")
	}

	list := val.Data.([]string)
	if len(list) == 0 {
		return "", errors.New("list is empty")
	}

	// get last element and remove it
	lastIdx := len(list) - 1
	item := list[lastIdx]
	list = list[:lastIdx]

	// if list is empty after pop, delete the key
	if len(list) == 0 {
		delete(db.data, key)
	} else {
		val.Data = list
		db.data[key] = val
	}

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("RPOP", key); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return item, nil
}

// LRange returns a range of elements from a list
func (db *FlexDB) LRange(key string, start, stop int) ([]string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return []string{}, nil // Redis returns empty list if key doesn't exist
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return []string{}, nil
	}

	if val.Type != TypeList {
		return nil, errors.New("value is not a list")
	}

	list := val.Data.([]string)
	length := len(list)

	// handle negative indices (counting from the end)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// boundary checks
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	// empty range
	if start > stop || start >= length {
		return []string{}, nil
	}

	return list[start : stop+1], nil
}

// LLen returns the length of a list
func (db *FlexDB) LLen(key string) (int, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return 0, nil // Redis returns 0 if key doesn't exist
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return 0, nil
	}

	if val.Type != TypeList {
		return 0, errors.New("value is not a list")
	}

	list := val.Data.([]string)
	return len(list), nil
}

// LIndex returns an element from a list by its index
func (db *FlexDB) LIndex(key string, index int) (string, error) {
	db.lock.RLock()
	defer db.lock.RUnlock()

	val, exists := db.data[key]
	if !exists {
		return "", errors.New("key not found")
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return "", errors.New("key not found")
	}

	if val.Type != TypeList {
		return "", errors.New("value is not a list")
	}

	list := val.Data.([]string)
	length := len(list)

	// handle negative index
	if index < 0 {
		index = length + index
	}

	// boundary check
	if index < 0 || index >= length {
		return "", errors.New("index out of range")
	}

	return list[index], nil
}

// LSet sets the value of an element in a list by its index
func (db *FlexDB) LSet(key string, index int, value string) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return errors.New("key not found")
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return errors.New("key not found")
	}

	if val.Type != TypeList {
		return errors.New("value is not a list")
	}

	list := val.Data.([]string)
	length := len(list)

	// handle negative index
	if index < 0 {
		index = length + index
	}

	// boundary check
	if index < 0 || index >= length {
		return errors.New("index out of range")
	}

	// set the value
	list[index] = value
	val.Data = list
	db.data[key] = val

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("LSET", key, fmt.Sprintf("%d", index), value); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return nil
}

// LRem removes elements from a list
func (db *FlexDB) LRem(key string, count int, value string) (int, error) {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return 0, nil // Redis returns 0 if key doesn't exist
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		return 0, nil
	}

	if val.Type != TypeList {
		return 0, errors.New("value is not a list")
	}

	list := val.Data.([]string)
	removed := 0

	// count > 0: remove from head to tail
	// count < 0: remove from tail to head
	// count = 0: remove all occurrences
	if count >= 0 {
		// remove from head to tail
		for i := 0; i < len(list) && (count == 0 || removed < count); {
			if list[i] == value {
				list = append(list[:i], list[i+1:]...)
				removed++
			} else {
				i++
			}
		}
	} else {
		// remove from tail to head
		for i := len(list) - 1; i >= 0 && (count == 0 || removed < -count); {
			if list[i] == value {
				list = append(list[:i], list[i+1:]...)
				removed++
			}
			i--
		}
	}

	// if list is empty after removal, delete the key
	if len(list) == 0 {
		delete(db.data, key)
	} else {
		val.Data = list
		db.data[key] = val
	}

	// Log AOF if enabled and elements were removed
	if removed > 0 && db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("LREM", key, fmt.Sprintf("%d", count), value); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	if removed > 0 {
		db.triggerWrite()
	}

	return removed, nil
}

// LTrim trims a list to the specified range
func (db *FlexDB) LTrim(key string, start, stop int) error {
	db.lock.Lock()
	defer db.lock.Unlock()

	val, exists := db.data[key]
	if !exists {
		return nil // Redis returns OK if key doesn't exist
	}

	// check if key has expired
	if val.Expiration != nil && time.Now().After(*val.Expiration) {
		delete(db.data, key)
		return nil
	}

	if val.Type != TypeList {
		return errors.New("value is not a list")
	}

	list := val.Data.([]string)
	length := len(list)

	// handle negative indices
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// boundary checks
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}

	// empty range
	if start > stop || start >= length {
		// delete the key if the range is empty
		delete(db.data, key)
	} else {
		// trim the list
		list = list[start : stop+1]
		val.Data = list
		db.data[key] = val
	}

	// Log AOF if enabled
	if db.aof != nil && db.aof.enabled {
		if err := db.aof.LogCommand("LTRIM", key, fmt.Sprintf("%d", start), fmt.Sprintf("%d", stop)); err != nil {
			fmt.Printf("Error logging to AOF: %v\n", err)
		}
	}

	db.triggerWrite()
	return nil
}