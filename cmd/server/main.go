package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"flex-db/internal/db"
	"flex-db/internal/protocol"
)

func main() {
	// Command line flags
	port := flag.Int("port", 9000, "Port to listen on")
	dbFile := flag.String("db", "data.json", "Database file path")

	// AOF configuration
	enableAOF := flag.Bool("aof", false, "Enable persistence")
	aofFile := flag.String("aof-file", "flexdb.aof", "AOF file path")
	aofSyncPolicy := flag.String("aof-sync", "everySec", "AOF sync policy: always, everySec, no")
	flag.Parse()

	//add AOF options if enabled
	var options []db.Option

	if *enableAOF {
		var syncPolicy db.AOFSyncPolicy

		switch *aofSyncPolicy {
		case "always":
			syncPolicy = db.AOFSyncAlways
		case "everysec", "everySec": 
			syncPolicy = db.AOFSyncEverySecond
		case "no":
			syncPolicy = db.AOFSyncNever
		default:
			fmt.Printf("Invalid AOF sync policy: %s, using 'everySec'\n", *aofSyncPolicy)
			syncPolicy = db.AOFSyncEverySecond
		}
		
		options = append(options, db.WithAOF(*aofFile, syncPolicy))
		fmt.Printf("AOF persistence enabled with file: %s, sync policy: %s\n", *aofFile, *aofSyncPolicy)
	}

	// Initialize database
	database := db.NewFlexDB(*dbFile, options...)
	handler := protocol.NewHandler(database)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("FlexDB server started on port %d\n", *port)

	// Handle connections in a separate goroutine
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				// Check if server is shutting down
				select {
				case <-sigChan:
					return
				default:
					fmt.Println("Connection error:", err)
					continue
				}
			}

			go handler.HandleConnection(conn)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	fmt.Println("\nShutting down server...")
	listener.Close()
	database.Flush()
	fmt.Println("Server shutdown complete")
} 