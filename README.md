# WabiSaby Plugins

This repository contains the source code for all official WabiSaby plugins. Plugins extend the functionality of the WabiSaby platform by providing additional capabilities such as storage providers, content resolvers, and more.

## Repository Structure

Each plugin is organized in its own directory following Go best practices:

```
plugin-name/
├── cmd/
│   └── plugin/
│       └── main.go          # Plugin entry point
├── internal/
│   ├── package1/            # Internal packages
│   └── package2/
├── manifest.yaml            # Plugin metadata
├── go.mod                   # Go module definition
└── README.md                # Plugin-specific documentation
```

## Building Plugins

### Prerequisites

- Go 1.24.0 or later
- Access to the WabiSaby-Go repository (for SDK dependencies)
- The WabiSaby-Go repository should be cloned at the same level as this repository:
  ```
  coding/
  ├── WabiSaby-Go/
  └── WabiSaby-Plugins/
  ```

### Build All Plugins

To build all plugins in the repository:

```bash
./build.sh
```

### Build a Specific Plugin

To build a specific plugin:

```bash
./build.sh wabisaby-network-storage
```

The build script will:
1. Read the plugin's `manifest.yaml` to extract version information
2. Download and tidy Go module dependencies
3. Build the plugin binary from `cmd/plugin/main.go`
4. Output the binary as `plugin` in the plugin directory

## Plugin Development

### Creating a New Plugin

1. Create a new directory for your plugin:
   ```bash
   mkdir -p my-plugin/cmd/plugin my-plugin/internal
   ```

2. Create the plugin structure:
   - `cmd/plugin/main.go` - Entry point that calls `sdk.Serve()`
   - `internal/` - Your plugin implementation
   - `manifest.yaml` - Plugin metadata (id, name, version, permissions, etc.)
   - `go.mod` - Go module with dependencies

3. Implement the plugin interfaces:
   - `sdk.Plugin` - Base interface (Initialize, Shutdown)
   - `sdk.CommandPlugin` - For command-based plugins
   - `sdk.StorageProvider` - For storage provider plugins
   - `sdk.MetadataResolver` - For content resolver plugins
   - `sdk.ContentDownloader` - For download plugins

4. Add your plugin to this repository and create a pull request

### Plugin Manifest

Each plugin must have a `manifest.yaml` file with the following structure:

```yaml
id: my-plugin
name: My Plugin
description: Description of what the plugin does
version: 1.0.0
author: Your Name
runtime: go
permissions:
  - storage:read
  - storage:write
  - http:fetch
official: true  # Set to false for community plugins
```

### Dependencies

Plugins depend on the WabiSaby SDK which is located in the WabiSaby-Go repository. The `go.mod` file should use a `replace` directive to point to the local WabiSaby-Go repository:

```go
module github.com/wabisaby/wabisaby-plugins/my-plugin

go 1.24.0

require (
    github.com/wabisaby/wabisaby v0.0.0
    // other dependencies...
)

replace github.com/wabisaby/wabisaby => ../../../WabiSaby-Go
```

## Available Plugins

### wabisaby-network-storage

Distributed storage provider powered by WabiSaby Community Nodes (IPFS). Allows users to store audio files on the distributed network and get paid for providing storage.

See [wabisaby-network-storage/README.md](wabisaby-network-storage/README.md) for more details.

## Contributing

1. Fork this repository
2. Create a feature branch
3. Implement your plugin following the structure and conventions
4. Add tests if applicable
5. Update documentation
6. Submit a pull request

## License

Copyright (c) 2026 WabiSaby. All rights reserved.

This source code is proprietary and confidential. Unauthorized copying, modification, distribution, or use of this software, via any medium is strictly prohibited without the express written permission of WabiSaby.
