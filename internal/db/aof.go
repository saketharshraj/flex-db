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
		}
	}

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