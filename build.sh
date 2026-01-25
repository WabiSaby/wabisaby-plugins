#!/bin/bash

# Build script for WabiSaby Plugins
# Usage: ./build.sh [plugin-name]
# If no plugin name is provided, builds all plugins

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to build a single plugin
build_plugin() {
    local plugin_dir="$1"
    local plugin_name=$(basename "$plugin_dir")
    
    if [ ! -d "$plugin_dir" ]; then
        echo -e "${RED}Error: Plugin directory not found: $plugin_dir${NC}"
        return 1
    fi
    
    if [ ! -f "$plugin_dir/manifest.yaml" ]; then
        echo -e "${YELLOW}Warning: No manifest.yaml found in $plugin_dir, skipping...${NC}"
        return 0
    fi
    
    echo -e "${GREEN}Building plugin: $plugin_name${NC}"
    
    cd "$plugin_dir"
    
    # Extract version from manifest.yaml
    local version=$(grep "^version:" manifest.yaml | awk '{print $2}' | tr -d '"' || echo "unknown")
    
    if [ "$version" = "unknown" ]; then
        echo -e "${YELLOW}Warning: Could not extract version from manifest.yaml${NC}"
    fi
    
    # Ensure go.mod exists
    if [ ! -f "go.mod" ]; then
        echo -e "${RED}Error: go.mod not found in $plugin_dir${NC}"
        cd "$SCRIPT_DIR"
        return 1
    fi
    
    # Download dependencies
    echo "  Downloading dependencies..."
    go mod download || {
        echo -e "${RED}Error: Failed to download dependencies${NC}"
        cd "$SCRIPT_DIR"
        return 1
    }
    
    # Tidy dependencies
    go mod tidy || {
        echo -e "${YELLOW}Warning: go mod tidy had issues${NC}"
    }
    
    # Build the plugin
    echo "  Building binary..."
    if go build -o plugin ./cmd/plugin; then
        echo -e "${GREEN}  ✓ Successfully built $plugin_name (version: $version)${NC}"
        echo "  Binary location: $plugin_dir/plugin"
    else
        echo -e "${RED}  ✗ Failed to build $plugin_name${NC}"
        cd "$SCRIPT_DIR"
        return 1
    fi
    
    cd "$SCRIPT_DIR"
    return 0
}

# Main execution
if [ $# -eq 0 ]; then
    # Build all plugins
    echo -e "${GREEN}Building all plugins...${NC}"
    echo ""
    
    for plugin_dir in */; do
        # Skip if not a directory or if it's a hidden directory
        [ -d "$plugin_dir" ] || continue
        [[ "$plugin_dir" == .* ]] && continue
        
        # Check if it has a manifest.yaml (indicates it's a plugin)
        if [ -f "$plugin_dir/manifest.yaml" ]; then
            build_plugin "$plugin_dir"
            echo ""
        fi
    done
    
    echo -e "${GREEN}Build process completed!${NC}"
else
    # Build specific plugin
    plugin_name="$1"
    plugin_dir="$plugin_name"
    
    if [ ! -d "$plugin_dir" ]; then
        echo -e "${RED}Error: Plugin '$plugin_name' not found${NC}"
        exit 1
    fi
    
    build_plugin "$plugin_dir"
fi
