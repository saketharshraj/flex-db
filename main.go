package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	db := NewFlexDB("data.json")
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to FlexDB CLI!")
	fmt.Println("Type 'help' for commands.")

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		args := strings.SplitN(input, " ", 3)

		if len(args) == 0 {
			continue
		}

		command := strings.ToLower(args[0])

		switch command {
		case "set":
			if len(args) < 3 {
				fmt.Println("Usage: set <key> <value>")
				continue
			}
			db.Set(args[1], args[2])
			fmt.Println("OK")

		case "get":
			if len(args) < 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			val, err := db.Get(args[1])
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println(val)
			}

		case "delete":
			if len(args) < 2 {
				fmt.Println("Usage: delete <key>")
				continue
			}
			err := db.Delete(args[1])
			if err != nil {
				fmt.Println("Error:", err)
			} else {
				fmt.Println("Deleted")
			}

		case "all":
			all := db.All()
			for k, v := range all {
				fmt.Printf("%s: %s\n", k, v)
			}

		case "help":
			fmt.Println("Available commands:")
			fmt.Println(" set <key> <value>  - Store a key-value pair")
			fmt.Println(" get <key>          - Get value by key")
			fmt.Println(" delete <key>       - Delete key-value pair")
			fmt.Println(" all                - Show all data")
			fmt.Println(" exit               - Quit FlexDB")

		case "exit":
			fmt.Println("Bye 👋")
			return

		default:
			fmt.Println("Unknown command. Type 'help' to see available commands.")
		}
	}
}
