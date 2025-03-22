package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// handleConnection runs in its own goroutine per client
func handleConnection(conn net.Conn, db *FlexDB) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	for {
		writer.WriteString("> ")
		writer.Flush()

		// Read client input
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected")
			return
		}
		line = strings.TrimSpace(line)
		args := strings.SplitN(line, " ", 3)
		if len(args) == 0 || args[0] == "" {
			continue
		}

		cmd := strings.ToUpper(args[0])

		switch cmd {
		case "SET":
			if len(args) < 3 {
				writer.WriteString("ERROR: Usage SET <key> <value>\n")
			} else {
				db.Set(args[1], args[2])
				writer.WriteString("OK\n")
			}

		case "GET":
			if len(args) < 2 {
				writer.WriteString("ERROR: Usage GET <key>\n")
			} else {
				val, err := db.Get(args[1])
				if err != nil {
					writer.WriteString("ERROR: key not found\n")
				} else {
					writer.WriteString(val + "\n")
				}
			}

		case "DELETE":
			if len(args) < 2 {
				writer.WriteString("ERROR: Usage DELETE <key>\n")
			} else {
				err := db.Delete(args[1])
				if err != nil {
					writer.WriteString("ERROR: key not found\n")
				} else {
					writer.WriteString("OK\n")
				}
			}

		case "ALL":
			all := db.All()
			for k, v := range all {
				writer.WriteString(fmt.Sprintf("%s: %s\n", k, v))
			}
			writer.WriteString("END\n")

		case "EXIT":
			writer.WriteString("Bye 👋\n")
			writer.Flush()
			return

		default:
			writer.WriteString("ERROR: Unknown command\n")
		}
		writer.Flush()
	}
}
