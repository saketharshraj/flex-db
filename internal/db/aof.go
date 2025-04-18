package db

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AOFSyncPolicy determines when to sync AOF to disk
type AOFSyncPolicy int

type AOFPersistence struct {
	db         *FlexDB
	file       *os.File
	writer     *bufio.Writer
	filePath   string
	mu         sync.Mutex
	enabled    bool
	syncPolicy AOFSyncPolicy
}

const (
	// AOFSyncAlways syncs after every write
	AOFSyncAlways AOFSyncPolicy = iota
	// AOFSyncEverySecond sync once per second
	AOFSyncEverySecond
	// AOFSyncNever lets the OS handle syncing
	AOFSyncNever
)

// To create a new AOF p ersistence manager
func NewAOFPersistence(db *FlexDB, filePath string, syncPolicy AOFSyncPolicy) (*AOFPersistence, error) {
	aof := &AOFPersistence{
		db:         db,
		filePath:   filePath,
		syncPolicy: syncPolicy,
		enabled:    true,
	}

	// create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.Mkdir(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for AOF: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open AOF file: %w", err)
	}

	aof.file = file
	aof.writer = bufio.NewWriter(file)

	// start background sync if using every-second policy
	if syncPolicy == AOFSyncEverySecond {
		go aof.backgroundSync()
	}

	return aof, nil
}


func (aof *AOFPersistence) sync() error {
	if err := aof.writer.Flush(); err != nil {
		return err
	}

	return aof.file.Sync()
}


func (aof *AOFPersistence) backgroundSync() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		aof.mu.Lock()
		aof.sync()
		aof.mu.Unlock()
	}
}
