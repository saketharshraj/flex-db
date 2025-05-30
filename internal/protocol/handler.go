package protocol

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"flex-db/internal/db"
)

// Handler manages client connections
type Handler struct {
	DB *db.FlexDB
	registry *CommandRegistry
}

// NewHandler creates a new command handler
func NewHandler(database *db.FlexDB) *Handler {
	return &Handler{
		DB: database,
		registry: NewCommandRegistry(),
	}
}

func validateArgs(cmd string, args []string, expected int) bool {
	if len(args) < expected {
		fmt.Printf("%s requires %d arguments\n", cmd, expected)
		return false
	}
	return true
}



func (h *Handler) HandleConnection(conn net.Conn) {
	protocolType, reader, err := DetectProtocol(conn)
	if err != nil {
		fmt.Printf("Error detecting protocol: %v\n", err)
		conn.Close()
		return
	}

	switch protocolType {
	case RESPProtocol:
		h.HandleRESPConnection(conn, reader)
	case TextProtocol:
		h.HandleTextConnection(conn, reader)
	}
}

// HandleConnection processes client commands
func (h *Handler) HandleTextConnection(conn net.Conn, reader *bufio.Reader) {
	defer conn.Close()
	addr := conn.RemoteAddr().String()
	fmt.Printf("[+] Client connected: %s\n", addr)
	defer fmt.Printf("[-] Client disconnected: %s\n", addr)

	writer := bufio.NewWriter(conn)

	for {
		writer.WriteString("> ")
		writer.Flush()

		// Read client input
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		args := strings.SplitN(line, " ", 4) // Allow for TTL as 4th argument
		if len(args) == 0 || args[0] == "" {
			continue
		}

		cmd := strings.ToUpper(args[0])
		switch cmd {
		case "SET":
			if !validateArgs(cmd, args, 2) {
				writer.WriteString("SET command requires at least two arguments\n")
				continue
			}
			key := args[1]
			value := args[2]
			
			var expiry *time.Time
			if len(args) == 4 {
				seconds, err := strconv.ParseInt(args[3], 10, 64)
				if err != nil {
					writer.WriteString("Invalid expiration format\n")
					continue
				}
				t := time.Now().Add(time.Duration(seconds) * time.Second)
				expiry = &t
			}
			
			h.DB.Set(key, value, expiry)
			writer.WriteString("OK\n")
		case "GET":
			if !validateArgs(cmd, args, 2) {
				writer.WriteString("GET command requires one argument\n")
				continue
			}
			key := args[1]
			value, err := h.DB.Get(key)
			if err != nil {
				writer.WriteString("(nil)\n")
			} else {
				writer.WriteString(fmt.Sprintf("%v\n", value))
			}
		case "ALL":
			all := h.DB.All()
			for k, v := range all {
				writer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			writer.WriteString("END\n")
		case "DEL":
			if !validateArgs(cmd, args, 2) {
				writer.WriteString("DEL command requires at least one argument\n")
				continue
			}
			for _, key := range args[1:] {
				h.DB.Delete(key)
			}
			writer.WriteString("OK\n")
		case "EXPIRE":
			if !validateArgs(cmd, args, 3) {
				writer.WriteString("EXPIRE command requires two arguments\n")
				continue
			}
			key := args[1]
			duration, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				writer.WriteString("Invalid duration format\n")
				continue
			}
			h.DB.Expire(key, time.Duration(duration)*time.Second)
			writer.WriteString("OK\n")
		case "TTL":
			if !validateArgs(cmd, args, 2) {
				writer.WriteString("TTL command requires one argument\n")
				continue
			}
			key := args[1]
			duration, err := h.DB.TTL(key)
			if err != nil {
				writer.WriteString("-1\n")
			} else {
				writer.WriteString(fmt.Sprintf("%.0f\n", duration.Seconds()))
			}
		case "FLUSH":
			h.DB.Flush()
			writer.WriteString("OK\n")
		
		case "BGREWRITE":
			go func() {
				if err := h.DB.RewriteAOF(); err != nil {
					fmt.Printf("Error rewriting AOF: %v\n", err)
				}
			}()
			writer.WriteString("Background rewrite started\n")
		
		case "HELP":
			writer.WriteString("Available commands:\n\n")
			for _, cmd := range AVAILABLE_COMMANDS {
				writer.WriteString(fmt.Sprintf("  %s\n", cmd))
			}
			writer.WriteString("\n")
		case "EXIT":
			writer.WriteString("Bye 👋\n")
			writer.Flush()
			h.DB.Flush()
			return
		default:
			writer.WriteString("Unknown command\n")
		}
	}
}