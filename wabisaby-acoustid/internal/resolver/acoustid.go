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

package resolver

import (
	"context"
	"fmt"
	"os"
	"strings"

	sdk "github.com/wabisaby/wabisaby-plugin-sdk"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-acoustid/internal/acoustid"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-acoustid/internal/fingerprint"
)

const acoustidAPIKeyEnv = "ACOUSTID_API_KEY"

// AcoustIDPlugin identifies uploaded songs using audio fingerprinting.
// It uses Chromaprint (fpcalc) to generate fingerprints and the AcoustID
// API to look up matching MusicBrainz recording IDs.
type AcoustIDPlugin struct {
	*sdk.MetadataResolverPlugin
	acoustidClient *acoustid.Client
}

func NewAcoustIDPlugin() *AcoustIDPlugin {
	apiKey := os.Getenv(acoustidAPIKeyEnv)
	return &AcoustIDPlugin{
		MetadataResolverPlugin: sdk.NewMetadataResolverPlugin(),
		acoustidClient:         acoustid.NewClient(apiKey),
	}
}

// ResolveURL accepts a local file path and identifies the audio via fingerprinting.
func (p *AcoustIDPlugin) ResolveURL(ctx *sdk.Context, req *sdk.ResolveURLRequest) (*sdk.ResolveResult, error) {
	filePath := req.URL

	if !isLocalFilePath(filePath) {
		return nil, fmt.Errorf("AcoustID only handles local file paths")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if !fingerprint.IsAvailable() {
		return nil, fmt.Errorf("fpcalc not installed")
	}

	fp, err := fingerprint.Generate(context.Background(), filePath)
	if err != nil {
		return nil, fmt.Errorf("generate fingerprint: %w", err)
	}

	result, err := p.acoustidClient.Lookup(context.Background(), fp.Fingerprint, fp.Duration)
	if err != nil {
		return nil, fmt.Errorf("acoustid lookup: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	metadata := &sdk.SongMetadata{
		Title: result.Title,
	}
	if result.ArtistName != "" {
		metadata.Artist = &result.ArtistName
	}
	duration := fp.Duration
	metadata.Duration = &duration

	return &sdk.ResolveResult{
		Metadata: metadata,
	}, nil
}

// Search is not supported by AcoustID - it only does fingerprint lookups.
func (p *AcoustIDPlugin) Search(ctx *sdk.Context, req *sdk.SearchRequest) ([]*sdk.SearchResult, error) {
	return nil, nil
}

// CanHandle returns true for local file paths.
func (p *AcoustIDPlugin) CanHandle(url string) bool {
	return isLocalFilePath(url)
}

// SupportedDomains returns empty since AcoustID handles local files, not URLs.
func (p *AcoustIDPlugin) SupportedDomains() []string {
	return []string{}
}

// IdentifyResult contains the result of audio fingerprint identification.
type IdentifyResult struct {
	RecordingID string
	Title       string
	ArtistName  string
	ArtistID    string
	Score       float64
	Duration    int
}

// Identify runs fingerprinting on a local audio file and returns the best match.
// This is the primary command exposed for identifying uploaded songs.
func (p *AcoustIDPlugin) Identify(filePath string) (*IdentifyResult, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", filePath)
	}

	if !fingerprint.IsAvailable() {
		return nil, fmt.Errorf("fpcalc not installed - cannot generate fingerprint")
	}

	fp, err := fingerprint.Generate(context.Background(), filePath)
	if err != nil {
		return nil, fmt.Errorf("generate fingerprint: %w", err)
	}

	result, err := p.acoustidClient.Lookup(context.Background(), fp.Fingerprint, fp.Duration)
	if err != nil {
		return nil, fmt.Errorf("acoustid lookup: %w", err)
	}

	if result == nil {
		return nil, nil
	}

	return &IdentifyResult{
		RecordingID: result.RecordingID,
		Title:       result.Title,
		ArtistName:  result.ArtistName,
		ArtistID:    result.ArtistID,
		Score:       result.Score,
		Duration:    fp.Duration,
	}, nil
}

func isLocalFilePath(path string) bool {
	return strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../")
}
