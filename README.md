# 🚀 FlexDB

**FlexDB** is a lightweight, in-memory NoSQL key-value store written in pure Go, with a TCP interface and persistent disk storage. It's designed for learning, benchmarking, and exploring what it takes to build a database from scratch.

## 💡 Features

- 🧠 In-memory key-value storage using Go maps
- 📦 Persistent disk writes using JSON
- ⏳ Buffered write queue to reduce I/O overhead
- ⚡ Fast and concurrent TCP server
- 🛠 Minimal external dependencies (zero libraries used)
- 📊 Benchmarking suite to test read/write performance under load

## 📁 Project Structure

```
flexdb/
├── db.go              # In-memory DB logic and persistence
├── main.go            # TCP server entry point
├── server.go          # Client connection and command handling
benchmark/
└── ...                # Benchmarking tools (in separate directory)
```

## 📌 How It Works

1. **Data Storage:** Key-value pairs are stored in RAM using a `map[string]string`.
2. **Persistence:** The entire dataset is periodically flushed to a `data.json` file.
3. **Write Optimization:** A background goroutine batches disk writes using a write queue, reducing I/O strain.
4. **TCP Interface:** Clients can connect via `telnet` or `nc` and issue commands like `SET`, `GET`, `DELETE`, etc.

## 🧪 Supported Commands

Once connected to the FlexDB server over TCP:

| Command              | Description                            |
|----------------------|----------------------------------------|
| `SET key value`      | Store a value for the given key        |
| `GET key`            | Retrieve value for a key               |
| `DELETE key`         | Remove a key-value pair                |
| `ALL`                | Dump all key-value pairs               |
| `EXIT`               | Close the connection                   |

Example session:
```
$ nc localhost 9000
> SET name harsh
OK
> GET name
harsh
> DELETE name
OK
> EXIT
Bye 👋
```

## 🚀 Getting Started

### ✅ Build and Run the Server

```bash
go build -o flexdb_server
./flexdb_server
```

The server listens on port `9000` by default.

### 🔌 Connect via Terminal

```bash
nc localhost 9000
```

## 📈 Performance Benchmarks

FlexDB includes a benchmarking tool (`/benchmark`) that tests single and multi-client performance under various loads.

Here are the results:

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

## 🧱 How Persistence Works

- All `SET` and `DELETE` operations are first written to RAM.
- A **background write loop** runs every 2 seconds (or on batched triggers) to save the current state to `data.json`.
- This prevents blocking the main flow on every disk write.

## 🧰 Developer Notes

- No external database or storage engine used.
- Entire system built from scratch for educational value.
- Read/write operations are thread-safe via `sync.RWMutex`.

## 📅 Roadmap

- [x] In-memory key-value store
- [x] TCP server for client interaction
- [x] Buffered write system for persistence
- [x] Multi-client concurrency support
- [x] Benchmarking framework
- [ ] Append-only log (AOF) for better persistence
- [ ] TTL (time-to-live) support for keys
- [ ] Pub/Sub or Watchers
- [ ] CLI tool with command history
- [ ] Web dashboard for stats & inspection

## 👨‍💻 Author

Built by [Harsh Raj](https://github.com/saketharshraj) — for learning and fun.  
Feel free to fork, extend, and break things. ⚙️

## 📜 License

MIT — do whatever you want, just give credit if you learned something.