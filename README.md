# mCHAT

A chat application written in Go using mDNS (multicast DNS) to send encrypted messages within local networks, enabling peer-to-peer communication without central servers.

## How It Works

mCHAT leverages multicast DNS (mDNS) for peer discovery and secure message exchange:

1. **Zero Configuration**: Using mDNS, devices automatically discover each other on the local network without manual IP configuration
2. **End-to-End Encryption**: All messages are encrypted using AES-256 with unique session keys
3. **Serverless Design**: Direct peer-to-peer communication eliminates the need for central servers
4. **Local Network Focus**: Perfect for LANs, office environments, or home networks

## Getting Started

### Prerequisites
- Go 1.19 or higher
- Make

### Installation
1. Clone the repository:
```bash
git clone https://github.com/yourusername/mCHAT.git
cd mCHAT
```

2. Build the application:
```bash
make build
```

### Usage
1. Start the application:
```bash
make run
```

2. The application will automatically:
   - Generate encryption keys
   - Start discovering peers on your network
   - Open a chat interface

## Building

This project uses Make for building. The following commands are available:

```bash
make build        # Build for your current platform
make build-all    # Build for all platforms (Windows and Linux)
make build-linux  # Build for Linux
make build-windows # Build for Windows
make clean        # Clean build artifacts
make test        # Run tests
make run         # Build and run the application
```

## Technical Details

### mDNS Discovery
mCHAT uses multicast DNS for automatic peer discovery, similar to how printers and smart devices are discovered on local networks. This enables:
- Automatic peer detection without configuration
- Network isolation for security
- Resilience through decentralization

### Security
- All messages are encrypted using AES-256-GCM

## License

MIT
