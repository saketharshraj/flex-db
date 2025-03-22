package main

import (
	"fmt"
	"net"
)

func main() {
	db := NewFlexDB("data.json")

	listener, err := net.Listen("tcp", ":9000")
	if err != nil {
		panic(err)
	}
	fmt.Println("FlexDB server started on port 9000")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		addr := conn.RemoteAddr().String()
		fmt.Printf("[+] Client connected: %s\n", addr)

		go handleConnectionWithLogs(conn, db, addr)
	}
}
