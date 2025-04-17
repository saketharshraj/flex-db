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
	flag.Parse()

	// Initialize database
	database := db.NewFlexDB(*dbFile)
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