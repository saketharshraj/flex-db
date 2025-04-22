# 🚀 FlexDB

A lightweight Redis-like in-memory database with persistence, written in Go.


## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/saketharshraj/flex-db.git
cd flex-db

# Build the server
go build -o flexdb cmd/server/main.go
```

### Running the Server

```bash
# Run with default settings (port 9000, data.json)
./flexdb

# Run with custom settings
./flexdb --port 8000 --db custom_data.json

# Run with AOF enabled
./flexdb --aof --aof-file data.aof --aof-sync everysec

# Run with custom settings and AOF
./flexdb --port 8000 --db custom_data.json --aof --aof-file custom.aof --aof-sync always
```

### Connecting to FlexDB

You can use any TCP client like `telnet` or `nc` (netcat):

```bash
nc localhost 9000
```

### Example Session (Text Protocol)

```
$ nc localhost 9000
> SET name harsh
OK
> GET name
harsh
> SET counter 10 60
OK
> TTL counter
60
> DEL name
OK
> GET name
(nil)
> FLUSH
OK
> EXIT
Bye 👋
```

### Example Session (RESP Protocol)

```
$ redis-cli -p 9000
127.0.0.1:9000> SET name "harsh"
OK
127.0.0.1:9000> GET name
"harsh"
127.0.0.1:9000> PING
PONG
```

## 💡 Features

- 🧠 In-memory key-value store with disk persistence
- 📦 Support for string data types (with plans for lists and hashes)
- ⏳ Key expiration (TTL) support
- ⚡ Dual protocol support: Simple text protocol and RESP (Redis protocol)
- 🔄 Protocol auto-detection for client compatibility
- 🛠 Atomic file operations for data safety
- 📊 Concurrent access with read-write locks
- 🔄 Background expiration checking
- 💾 Buffered write system for improved performance
- 📝 Append-Only File (AOF) persistence for improved durability

## 🧪 Supported Commands

| Command | Description |
|---------|-------------|
| `SET <key> <value> [expiry_seconds]` | Set a key-value pair with optional expiration |
| `GET <key>` | Retrieve value for a key |
| `DEL <key> [key2...]` | Remove one or more key-value pairs |
| `EXPIRE <key> <seconds>` | Set expiration on an existing key |
| `TTL <key>` | Get remaining time to live for a key in seconds |
| `ALL` | List all key-value pairs |
| `FLUSH` | Force write to disk |
| `BGREWRITE` | Rewrite the AOF file in the background |
| `PING` | Test connection (RESP protocol) |
| `HELP` | Show available commands |
| `EXIT` | Close the connection |

## 📌 How It Works

1. **Data Storage:** Key-value pairs are stored in RAM using Go's map structure
2. **Persistence Options:**
   - **JSON Snapshots:** The dataset is periodically flushed to a JSON file
   - **AOF Persistence:** Each write operation is logged to an append-only file
3. **Write Optimization:** A background goroutine batches disk writes using a write queue
4. **Expiration:** A background process checks for and removes expired keys
5. **Protocol Support:** Automatic detection between text and RESP protocols
6. **TCP Interface:** Clients connect via TCP and issue commands in either protocol

### Persistence

- **JSON Persistence:**
  - Data is stored in a JSON file specified at startup
  - Writes are batched and performed every 2 seconds automatically

- **AOF Persistence:**
  - Each write command is logged to an append-only file
  - Three sync policies available:
    - `always`: Sync after every write (safest, slowest)
    - `everysec`: Sync once per second (good balance)
    - `no`: Let the OS handle syncing (fastest, least safe)
  - AOF can be rewritten/compacted with the `BGREWRITE` command

## 🏗️ Architecture

FlexDB follows a modular architecture with clear separation of concerns:

### Core Components

- **Database Engine** (`internal/db/db.go`): Manages the in-memory data store with concurrent access
- **Persistence Layer** (`internal/db/persistence.go`): Handles saving and loading data to/from disk
- **Server** (`cmd/server/main.go`): TCP server that accepts client connections
- **Protocol Handler** (`internal/protocol/handler.go`): Detects and processes client protocols
- **RESP Implementation** (`internal/resp/resp.go`): Redis protocol parser and serializer

### Data Flow

1. Client connects to the TCP server
2. Protocol detector identifies the client protocol (text or RESP)
3. Appropriate handler processes incoming commands
4. Database engine performs operations on the in-memory store
5. Changes are queued for persistence
6. Background workers handle persistence and key expiration

### Concurrency Model

- Read operations use read locks for concurrent access
- Write operations use write locks to ensure data consistency
- Background goroutines handle periodic tasks without blocking the main flow

## 📁 Project Structure

```
flexdb/
├── cmd/
│   └── server/        # Server entry point
│       └── main.go
├── internal/
│   ├── db/            # Database implementation
│   │   ├── db.go
│   │   └── persistence.go
│   ├── protocol/      # Protocol handling
│   │   ├── detector.go
│   │   ├── handler.go
│   │   └── resp_handler.go
│   ├── resp/          # RESP protocol implementation
│   │   └── resp.go
│   └── server/        # Connection handling
├── data.json          # Default database file
├── go.mod             # Go module definition
└── README.md
```

## 🔧 Implementation Details

### Protocol Support

- **Text Protocol**: Simple line-based protocol for human interaction
- **RESP Protocol**: Redis Serialization Protocol for Redis client compatibility
- **Auto-detection**: Server automatically detects which protocol the client is using

### Persistence

- Data is stored in a JSON file specified at startup
- Writes are batched and performed:
  - Every 2 seconds automatically
  - When triggered by operations that modify data
  - When explicitly requested with the `FLUSH` command
- Atomic file operations prevent data corruption

### TTL (Time-To-Live)

- Keys can be set with an expiration time in seconds
- A background goroutine checks for expired keys every second
- The `TTL` command returns the remaining time in seconds

### Data Types

- Currently supports string values
- Internal architecture is designed to support additional types in the future

## 📈 Performance Benchmarks

FlexDB includes a benchmarking tool that tests single and multi-client performance under various loads.

### 🔹 Test: `100_10_100` (1,000 ops total)

| Type         | Total Ops | Duration (ns) | Ops/sec |
|--------------|-----------|----------------|----------|
| Single       | 1,000     | 771,764,701    | 1295.73  |
| Multi-client | 1,000     | 662,069,786    | 1510.41  |

### 🔹 Test: `10000_20_100` (30,000 ops total)

| Type         | Total Ops | Duration (ns)  | Ops/sec |
|--------------|-----------|-----------------|----------|
| Single       | 10,000    | 47.5s (47,512,207,315 ns) | 210.47  |
| Multi-client | 20,000    | 191.1s (191,110,188,480 ns) | 104.65  |

### 🔹 After Buffered Write Optimization

| Type         | Total Ops | Duration (ns)  | Ops/sec |
|--------------|-----------|----------------|----------|
| Single       | 10,000    | 46.8s (46,835,041,777 ns) | 213.51  |
| Multi-client | 20,000    | 114.7s (114,783,912,900 ns) | 174.24  |

✅ **Buffered writes** improved multi-client throughput by **~67%**.

## 📈 Performance Considerations

- Read operations use a read lock for concurrent access
- Write operations use a write lock to ensure data consistency
- Buffered writes improve performance by batching disk operations
- Expired keys are cleaned up in the background

## 🧰 Developer Notes

- No external dependencies are used for the core functionality
- The entire system is built from scratch for educational purposes
- All operations are thread-safe using appropriate locking mechanisms

## 📅 Roadmap

- [x] In-memory key-value store
- [x] TCP server for client interaction
- [x] Buffered write system for persistence
- [x] Multi-client concurrency support
- [x] TTL (time-to-live) support for keys
- [x] RESP (Redis protocol) support
- [x] Append-only log (AOF) for better persistence
- [ ] Support for additional data types (Lists, Hashes)
- [ ] Pub/Sub messaging system
- [ ] Authentication
- [ ] CLI tool with command history
- [ ] Web dashboard for stats & monitoring

## 👨‍💻 Author

Built by [Harsh Raj](https://github.com/saketharshraj) — for learning and fun.  
Feel free to fork, extend, and contribute!

## 📜 License

MIT — do whatever you want, just give credit if you learned something.
