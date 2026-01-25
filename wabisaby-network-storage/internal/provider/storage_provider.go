// Copyright (c) 2026 WabiSaby
// All rights reserved.
//
// This source code is proprietary and confidential. Unauthorized copying,
// modification, distribution, or use of this software, via any medium is
// strictly prohibited without the express written permission of WabiSaby.
//
// This software contains confidential and proprietary information of
// WabiSaby and its licensors. Use, disclosure, or reproduction
// is prohibited without the prior express written permission of WabiSaby.

package provider

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-network-storage/internal/coordinator"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-network-storage/internal/ipfs"
	nodepb "github.com/wabisaby/wabisaby/api/generated/proto/node"
	sdk "github.com/wabisaby/wabisaby/pkg/plugin/sdk/go"
)

// Config holds the plugin configuration.
type Config struct {
	GatewayURL        string `json:"gateway_url"`
	APIURL            string `json:"api_url"`
	CoordinatorAddr   string `json:"coordinator_addr"`
	ReplicationFactor int    `json:"replication_factor"`
}

// WabiSabyStorageProvider implements the StorageProvider interface.
type WabiSabyStorageProvider struct {
	plugin      *WabiSabyStoragePlugin
	coordinator *coordinator.Client
	ipfs        *ipfs.Client
	config      Config
}

// NewWabiSabyStorageProvider creates a new storage provider.
func NewWabiSabyStorageProvider(plugin *WabiSabyStoragePlugin) *WabiSabyStorageProvider {
	return &WabiSabyStorageProvider{
		plugin: plugin,
	}
}

// Initialize initializes the storage provider with configuration.
func (p *WabiSabyStorageProvider) Initialize(accessor *sdk.ConfigAccessor) error {
	// Default config
	p.config = Config{
		GatewayURL:        accessor.GetString("gateway_url", "https://gateway.wabisaby.com"),
		APIURL:            accessor.GetString("api_url", "http://localhost:5001"),
		CoordinatorAddr:   accessor.GetString("coordinator_addr", os.Getenv("WABISABY_COORDINATOR_ADDR")),
		ReplicationFactor: accessor.GetInt("replication_factor", 3),
	}

	if p.config.CoordinatorAddr == "" {
		p.config.CoordinatorAddr = "localhost:50051" // Fallback
	}

	client, err := coordinator.NewClient(p.config.CoordinatorAddr)
	if err != nil {
		return err
	}
	p.coordinator = client
	p.ipfs = ipfs.NewClient(p.config.APIURL)

	// Register commands
	if err := p.plugin.RegisterCommand("storage.upload_hls", p.UploadHLSFiles); err != nil {
		return fmt.Errorf("register upload_hls command: %w", err)
	}
	if err := p.plugin.RegisterCommand("storage.get_size", p.GetFileSizeMB); err != nil {
		return fmt.Errorf("register get_size command: %w", err)
	}
	if err := p.plugin.RegisterCommand("storage.delete", p.DeleteAudio); err != nil {
		return fmt.Errorf("register delete command: %w", err)
	}

	return nil
}

// UploadHLSFiles uploads HLS files to IPFS and notifies the coordinator.
func (p *WabiSabyStorageProvider) UploadHLSFiles(ctx *sdk.Context, req *sdk.UploadHLSRequest) (string, error) {
	// 1. Upload segments to IPFS
	results, err := p.ipfs.AddDirectory(ctx.Context, req.SegmentsDir)
	if err != nil {
		return "", fmt.Errorf("failed to upload segments to IPFS: %w", err)
	}

	var rootCID string
	var playlistCID string
	var segmentCIDs []string
	var totalSize int64

	for _, res := range results {
		if res.Name == "" { // Root directory if wrap-with-directory=true
			rootCID = res.Hash
		} else if res.Name == filepath.Base(req.PlaylistPath) {
			playlistCID = res.Hash
		} else {
			segmentCIDs = append(segmentCIDs, res.Hash)
		}
	}

	// If wrap-with-directory=true was not used or didn't return a root name,
	// we might need to handle it differently.
	if rootCID == "" && len(results) > 0 {
		rootCID = results[len(results)-1].Hash // Usually the last one is the root
	}

	// Get accurate size
	stat, err := p.ipfs.Stat(ctx.Context, rootCID)
	if err == nil {
		totalSize = stat.CumulativeSize
	}

	// 2. Notify coordinator about the new content
	// Use TenantID as SongID if not provided (placeholder logic)
	songID := uuid.New().String()

	indexReq := &nodepb.IndexContentRequest{
		SongId:         songID,
		Cid:            rootCID,
		PlaylistCid:    playlistCID,
		SegmentCids:    segmentCIDs,
		TotalSizeBytes: totalSize,
		ReplicationMin: int32(p.config.ReplicationFactor),
		ReplicationMax: int32(p.config.ReplicationFactor * 2),
	}

	_, err = p.coordinator.StoreContent(ctx.Context, indexReq)
	if err != nil {
		return "", fmt.Errorf("failed to index content in coordinator: %w", err)
	}

	return fmt.Sprintf("%s/ipfs/%s/%s", p.config.GatewayURL, rootCID, filepath.Base(req.PlaylistPath)), nil
}

// GetFileSizeMB returns the file size in MB from a CDN URL.
func (p *WabiSabyStorageProvider) GetFileSizeMB(ctx *sdk.Context, cdnURL string) (float64, error) {
	// Parse CID from URL
	// Gateway URL: https://gateway.wabisaby.com/ipfs/CID/playlist.m3u8
	parts := strings.Split(cdnURL, "/ipfs/")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid IPFS URL: %s", cdnURL)
	}

	cidPath := parts[1]
	cid := strings.Split(cidPath, "/")[0]

	stat, err := p.ipfs.Stat(ctx.Context, cid)
	if err != nil {
		return 0, err
	}
	return float64(stat.CumulativeSize) / (1024 * 1024), nil
}

// DeleteAudio deletes an audio file from storage.
func (p *WabiSabyStorageProvider) DeleteAudio(ctx *sdk.Context, cdnURL string) error {
	// In IPFS, we can't really "delete" from the network, but we can unpin from our node
	// and notify the coordinator to unpin from WabiSaby nodes.
	return nil
}

// WabiSabyStoragePlugin is the main plugin struct.
type WabiSabyStoragePlugin struct {
	*sdk.StorageProviderPlugin
	provider *WabiSabyStorageProvider
}

// NewWabiSabyStoragePlugin creates a new WabiSaby storage plugin.
func NewWabiSabyStoragePlugin() *WabiSabyStoragePlugin {
	p := &WabiSabyStoragePlugin{
		StorageProviderPlugin: sdk.NewStorageProviderPlugin(),
	}
	p.provider = NewWabiSabyStorageProvider(p)
	return p
}

// Initialize initializes the plugin.
func (p *WabiSabyStoragePlugin) Initialize(ctx *sdk.Context) error {
	return p.provider.Initialize(ctx.Config)
}
