package protocol

var AVAILABLE_COMMANDS = []string{
	"SET key value [ttl]  - Set a key with optional TTL in seconds",
	"GET key              - Get value for a key",
	"DEL key              - Delete a key",
	"EXPIRE key seconds   - Set expiration time for a key",
	"TTL key              - Get remaining time for a key",
	"ALL                  - List all keys and values",
	"FLUSH                - Force save to disk",
	"BGREWRITE 			  - Rewrite the AOF file in the background", 
	"HELP                 - Show this help message",
	"EXIT                 - Close connection",
}