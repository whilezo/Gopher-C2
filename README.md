# Gopher-C2: A gRPC-Based Command & Control Framework

Developed as a final project following the completion of **Black Hat Go**, this framework is a high-performance, concurrent C2 system. It leverages gRPC for stealthy, efficient communication, featuring a multi-session server architecture and a robust administrative CLI.

## 🚀 Features

* **gRPC Protocol:** Uses HTTP/2 for encrypted, low-latency communication.
* **Hybrid Security Model:** * **mTLS (Admin -> Server):** Mutual authentication ensures only authorized operators can issue commands.
    * **Standard TLS (Implant -> Server):** Server-side encryption protects traffic without exposing client certificates in the implant binary.
* **Thread-Safe Session Management:** Custom `SessionManager` using `sync.RWMutex` to handle concurrent implant beacons without race conditions.
* **Individual Tasking:** Mailbox-style routing where every implant has its own dedicated work/result channels.
* **Parallel Broadcasting:** Send commands to all active implants simultaneously using Go's concurrency primitives (`sync.WaitGroup`).
* **Persistence:** SQLite database backend to track implant history, IP addresses, and status.
* **Identity via Metadata:** Uses gRPC metadata for implant identification, keeping the protocol clean and extensible.

---

## 🏗️ Architecture

- **Server:** Central hub managing the SQLite database and routing taskings. It exposes two separate gRPC listeners: one for untrusted Implants and one for the Admin client.
- **Admin CLI:** A subcommand-based tool built with `urfave/cli/v3` for managing the fleet.
- **Implant:** A lightweight Go binary designed to beacon to the server, execute shell instructions, and return results.

---

## 🛠️ Getting Started

### Prerequisites
* Go 1.21+
* Protoc compiler and Go gRPC plugins
* OpenSSL (for generating TLS certificates)

### 1. Build the Components
```bash
# Build everything at once
make build

# Build the Server
make server

# Build the Admin Tool
make admin

# Build the Implant
make implant
```

### 2. Launch the server
Ensure your certificates are in the directory
```bash
./server
```

## 🎮 Usage (Admin CLI)
The Admin tool uses subcommands for organized control:

### List implants
See which machines are online, offline, or recently registered.
```bash
./admin list
```

### Execute Command
Task a specific implant by its UUID.
```bash
./admin exec <implant-id> whoami /all
```

### Broadcast Command
Task every single registered implant at once.
```bash
./admin broadcast netstat -ano
```

### Delete Implant
Permanently remove an implant from the server and database.
```bash
./admin delete <implant-id>
```

---

### ⚖️ Disclaimer
This project was created for educational purposes and authorized security testing only. The author is not responsible for any misuse. Always obtain explicit permission before deploying security tools.

> Inspired by Black Hat Go
