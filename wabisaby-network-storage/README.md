# WabiSaby Network Storage Plugin

Distributed storage provider powered by WabiSaby Community Nodes (IPFS). This plugin enables users to store audio files on a distributed network and earn rewards for providing storage capacity.

## Features

- **Distributed Storage**: Upload HLS audio files to IPFS network
- **Automatic Replication**: Content is automatically replicated across WabiSaby nodes
- **Coordinator Integration**: Seamlessly integrates with WabiSaby Node Coordinator
- **CDN URLs**: Generates gateway URLs for accessing stored content

## Configuration

The plugin supports the following configuration options:

```json
{
  "gateway_url": "https://gateway.wabisaby.com",
  "api_url": "http://localhost:5001",
  "coordinator_addr": "localhost:50051",
  "replication_factor": 3
}
```

### Configuration Options

- **gateway_url** (string, default: `https://gateway.wabisaby.com`): The IPFS gateway URL for accessing content
- **api_url** (string, default: `http://localhost:5001`): The IPFS API endpoint
- **coordinator_addr** (string, default: `localhost:50051` or `WABISABY_COORDINATOR_ADDR` env var): The Node Coordinator gRPC address
- **replication_factor** (int, default: `3`): Minimum number of nodes to replicate content to

## Commands

The plugin provides the following commands:

### `storage.upload_hls`

Uploads HLS files (playlist and segments) to IPFS and registers them with the coordinator.

**Request:**
```json
{
  "playlist_path": "/path/to/playlist.m3u8",
  "segments_dir": "/path/to/segments",
  "base_filename": "song"
}
```

**Response:**
```json
{
  "cdn_url": "https://gateway.wabisaby.com/ipfs/QmXXX/playlist.m3u8"
}
```

### `storage.get_size`

Gets the total size of a file in MB from its CDN URL.

**Request:**
```json
{
  "cdn_url": "https://gateway.wabisaby.com/ipfs/QmXXX/playlist.m3u8"
}
```

**Response:**
```json
{
  "size_mb": 45.2
}
```

### `storage.delete`

Deletes an audio file from storage (unpins from IPFS nodes).

**Request:**
```json
{
  "cdn_url": "https://gateway.wabisaby.com/ipfs/QmXXX/playlist.m3u8"
}
```

## Architecture

The plugin is organized into the following packages:

- **cmd/plugin**: Entry point that initializes and serves the plugin
- **internal/coordinator**: gRPC client for communicating with the Node Coordinator
- **internal/ipfs**: HTTP client for IPFS API operations
- **internal/provider**: Main storage provider implementation

## Building

To build this plugin:

```bash
cd wabisaby-network-storage
go mod download
go build -o plugin ./cmd/plugin
```

Or use the repository build script:

```bash
./build.sh wabisaby-network-storage
```

## Dependencies

- `github.com/wabisaby/wabisaby-plugin-sdk` - WabiSaby Plugin SDK
- `github.com/WabiSaby/WabiSaby-Protos` - Node Coordinator protobuf definitions
- `github.com/google/uuid` - UUID generation library
- `google.golang.org/grpc` - gRPC client library
- `google.golang.org/protobuf` - Protocol buffers

## Usage

Once installed and enabled in WabiSaby, the plugin can be used as a storage provider. When audio files are transcoded to HLS format, they will be automatically uploaded to the IPFS network and registered with the Node Coordinator for replication across the WabiSaby network.

## Permissions

This plugin requires the following permission:
- `storage:provider` - Allows the plugin to act as a storage provider

## Version

Current version: 1.0.0