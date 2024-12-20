# MUX: Multi-User MUD Proxy

MUX is a lightweight and efficient TCP proxy for Multi-User Dungeon (MUD) servers. It acts as a middleman between clients and the MUD server, allowing multiple clients to connect to the same server instance. MUX maintains a persistent connection to the MUD server, even if no clients are connected, ensuring that no input from the server is missed.

## Features

- **Persistent MUD Connection**: Keeps the MUD connection active even when no clients are connected.
- **Client Broadcasting**: Forwards data from the MUD server to all connected clients.
- **Logging**: Configurable logging with support for log rotation using `lumberjack`.
- **Graceful Shutdown**: Cleans up resources and disconnects clients properly on termination.

## Requirements

- Go 1.16 or higher
- Internet access to connect to a MUD server

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/mux.git
   cd mux
   ```

2. Build the binary:
   ```bash
   go build -o mud-mux
   ```

3. Run the application:
   ```bash
   ./mudmux --mud <MUD_SERVER_ADDRESS> --local <LOCAL_ADDRESS> --log <LOG_FILE>
   ```

## Usage

### Command-Line Arguments

- `--mud`: Address of the MUD server in the format `host:port`.
- `--local`: Local address to listen on for client connections. Default: `:8888`.
- `--log`: Path to the log file. If omitted, logs are printed to stdout.

### Example

Run MUX with the following settings:
```bash
./mud-mux --mud "mud.example.com:4000" --local ":9000" --log "logs/mux.log"
```

### Connecting a Client

1. Start the MUX proxy.
2. Connect your MUD client to the local address (e.g., `localhost:9000`).
3. Interact with the MUD server through the proxy.

## How It Works

1. MUX connects to the specified MUD server and keeps the connection alive.
2. It listens for incoming client connections on the specified local address.
3. When a client connects:
   - Input from the client is forwarded to the MUD server.
   - Output from the MUD server is broadcasted to all connected clients.
4. If no clients are connected, MUX continues to read and log data from the MUD server.

## Graceful Shutdown

To stop the MUX server:

1. Press `Ctrl+C`.
2. MUX will:
   - Disconnect all clients.
   - Close the MUD server connection.
   - Clean up resources.

## Logging

MUX supports logging with optional rotation via the [lumberjack](https://github.com/natefinch/lumberjack) library. Logs include:
- Timestamps
- Connection events
- Data transfer between clients and the MUD server
- Errors during operation

### Log Rotation Settings

- **Max Size**: 500 MB
- **Max Backups**: 3
- **Max Age**: 28 days
- **Compression**: Enabled

## Contributing

Contributions are welcome! Feel free to submit issues, feature requests, or pull requests to improve MUX.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgments

- [lumberjack](https://github.com/natefinch/lumberjack) for log rotation.
