package db

import (
	"bufio"
	"flex-db/internal/utils"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

// To create a new AOF persistence manager
func NewAOFPersistence(db *FlexDB, filePath string, syncPolicy AOFSyncPolicy) (*AOFPersistence, error) {
	aof := &AOFPersistence{
		db:         db,
		filePath:   filePath,
		syncPolicy: syncPolicy,
		enabled:    true,
	}

	// create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if dir != "." {
		// Only try to create the directory if it's not the current directory
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory for AOF: %w", err)
		}
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

func (aof *AOFPersistence) LogCommand(cmd string, args ...string) error {
	if !aof.enabled {
		return nil
	}

	aof.mu.Lock()
	defer aof.mu.Unlock()

	// format the command before writing for AOF
	var sb strings.Builder
	sb.WriteString(cmd)
	for _, arg := range args {
		if strings.Contains(arg, " ") {
			sb.WriteString("\"")
			sb.WriteString(cmd)
			sb.WriteString("\"")
		} else {
			sb.WriteString(arg)
		}
	}
	sb.WriteString("\n")

	if _, err := aof.writer.WriteString(sb.String()); err != nil {
		return fmt.Errorf("failed to write to AOF buffer: %w", err)
	}

	if aof.syncPolicy == AOFSyncAlways {
		if err := aof.sync(); err != nil {
			return fmt.Errorf("failed to sync AOF: %w", err)
		}
	}

	return nil
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

func (aof *AOFPersistence) Close() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	if err := aof.sync();  err != nil {
		return err
	}

	return aof.file.Close()
}


func (aof *AOFPersistence) LoadAOF() error {
	// open file for reading
	file, err := os.Open(aof.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open AOF file for loading: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// parse the command
		parts, err := parseCommandLine(line)

		if err != nil {
			return fmt.Errorf("error in parsing AOF line: %w", err)
		}

		if len(parts) == 0 {
			continue
		}

		// execute the command
		cmd := strings.ToUpper(parts[0])
		args := parts[1:]

		switch cmd {
		case "SET":
			if len(args) < 2{
				continue
			}
			key := args[0]
			value := args[1]

			var expiry *time.Time
			if len(args) >= 3 {
				seconds, err := utils.ParseInt(args[2])
				if err == nil {
					t := time.Now().Add(time.Duration(seconds) * time.Second)
					expiry = &t
				}
			}
			aof.db.setWithoutLogging(key, value, expiry)
		case "EXPIRE":
			if len(args) != 2 {
				continue
			}

			key := args[0]
			seconds, err := utils.ParseInt(args[1])

			if err != nil {
				continue
			}

			aof.db.expireWithoutLogging(key, time.Duration(seconds)*time.Second)
		
		case "FLUSH":
			// no need for flush while replaying AOF
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning AOF file: %w", err)
	}

	return nil
}



// RewriteAOF compacts the AOF file by writing only commands needed for current state
func (aof *AOFPersistence) RewriteAOF() error {
	aof.mu.Lock()
	defer aof.mu.Unlock()

	tempFile := aof.filePath + ".rewrite"
	file, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary file for AOF rewrite: %w", err)
	}
	writer := bufio.NewWriter(file)

	snapshot := aof.db.All()

	// Write SET commands for all current keys
	for key, value := range snapshot {
		// Get TTL if any
		ttl, err := aof.db.TTL(key)
		var ttlArg string
		if err == nil && ttl > 0 {
			ttlArg = fmt.Sprintf(" %d", int(ttl.Seconds()))
		}

		cmd := fmt.Sprintf("SET %s %v%s\n", key, value, ttlArg)
		if _, err := writer.WriteString(cmd); err != nil {
			file.Close()
			return fmt.Errorf("failed to write to temporary AOF file: %w", err)
		}
	}

	// Flush and sync
	if err := writer.Flush(); err != nil {
		file.Close()
		return fmt.Errorf("failed to flush temporary AOF file: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("failed to sync temporary AOF file: %w", err)
	}
	file.Close()

	aof.sync()
	aof.file.Close()

	// Replace old file with new one
	if err := os.Rename(tempFile, aof.filePath); err != nil {
		return fmt.Errorf("failed to replace AOF file: %w", err)
	}

	// Reopen the AOF file
	file, err = os.OpenFile(aof.filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to reopen AOF file: %w", err)
	}

	aof.file = file
	aof.writer = bufio.NewWriter(file)

	return nil
}


func parseCommandLine(line string) ([]string, error) {
	var parts []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuotes = !inQuotes
		case c == ' ' && !inQuotes:
			if current.Len() > 0{
				parts = append(parts, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(c)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quotes in command")
	}

	return parts, nil

}