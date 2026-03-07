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
	"regexp"
	"strings"

	sdk "github.com/wabisaby/wabisaby-plugin-sdk"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-musicbrainz/internal/musicbrainz"
)

var recordingURLPattern = regexp.MustCompile(`musicbrainz\.org/recording/([a-f0-9-]{36})`)

type MusicBrainzPlugin struct {
	*sdk.MetadataResolverPlugin
	client *musicbrainz.Client
}

func NewMusicBrainzPlugin() *MusicBrainzPlugin {
	return &MusicBrainzPlugin{
		MetadataResolverPlugin: sdk.NewMetadataResolverPlugin(),
		client:                 musicbrainz.NewClient(),
	}
}

func (p *MusicBrainzPlugin) ResolveURL(ctx *sdk.Context, req *sdk.ResolveURLRequest) (*sdk.ResolveResult, error) {
	recordingID := extractRecordingID(req.URL)
	if recordingID == "" {
		return nil, fmt.Errorf("invalid MusicBrainz recording URL")
	}

	recording, err := p.client.GetRecording(context.Background(), recordingID)
	if err != nil {
		return nil, fmt.Errorf("get recording: %w", err)
	}
	if recording == nil {
		return nil, fmt.Errorf("recording not found")
	}

	metadata := recordingToMetadata(recording)

	return &sdk.ResolveResult{
		Metadata: metadata,
	}, nil
}

func (p *MusicBrainzPlugin) Search(ctx *sdk.Context, req *sdk.SearchRequest) ([]*sdk.SearchResult, error) {
	title, artist := parseQuery(req.Query)

	searchResult, err := p.client.SearchRecordings(context.Background(), title, artist)
	if err != nil {
		return nil, fmt.Errorf("search recordings: %w", err)
	}

	results := make([]*sdk.SearchResult, 0, len(searchResult.Recordings))
	for _, rec := range searchResult.Recordings {
		metadata := searchRecordingToMetadata(&rec)
		results = append(results, &sdk.SearchResult{
			Metadata: metadata,
			URL:      fmt.Sprintf("https://musicbrainz.org/recording/%s", rec.ID),
		})
	}

	return results, nil
}

func (p *MusicBrainzPlugin) CanHandle(url string) bool {
	return recordingURLPattern.MatchString(url)
}

func (p *MusicBrainzPlugin) SupportedDomains() []string {
	return []string{"musicbrainz.org"}
}

// GetRecordingByID fetches full recording details by MusicBrainz recording ID.
// This is a custom command exposed for enrichment workflows.
func (p *MusicBrainzPlugin) GetRecordingByID(recordingID string) (*RecordingDetails, error) {
	recording, err := p.client.GetRecording(context.Background(), recordingID)
	if err != nil {
		return nil, fmt.Errorf("get recording: %w", err)
	}
	if recording == nil {
		return nil, nil
	}

	details := &RecordingDetails{
		RecordingID: recording.ID,
		Title:       recording.Title,
	}

	if len(recording.ArtistCredit) > 0 {
		ac := recording.ArtistCredit[0]
		details.ArtistID = ac.Artist.ID
		details.ArtistName = ac.Artist.Name
		details.ArtistSortName = ac.Artist.SortName
	}

	if len(recording.Releases) > 0 {
		rel := recording.Releases[0]
		details.AlbumID = rel.ID
		details.AlbumTitle = rel.Title
		details.ReleaseYear = musicbrainz.ParseReleaseYear(rel.Date)

		if rel.ReleaseGroup != nil {
			details.ReleaseGroupID = rel.ReleaseGroup.ID
		}

		coverURL, _ := p.client.GetCoverArt(context.Background(), rel.ID)
		if coverURL != "" {
			details.CoverURL = &coverURL
		}
	}

	if len(recording.Tags) > 0 {
		genres := make([]string, 0, len(recording.Tags))
		for _, tag := range recording.Tags {
			genres = append(genres, tag.Name)
		}
		details.Genres = genres
	}

	if len(recording.ISRCs) > 0 {
		details.ISRC = &recording.ISRCs[0]
	}

	if recording.Length != nil {
		durationSec := *recording.Length / 1000
		details.Duration = &durationSec
	}

	return details, nil
}

// RecordingDetails contains enriched metadata from MusicBrainz.
type RecordingDetails struct {
	RecordingID    string
	Title          string
	ArtistID       string
	ArtistName     string
	ArtistSortName string
	AlbumID        string
	AlbumTitle     string
	ReleaseGroupID string
	ReleaseYear    *int
	CoverURL       *string
	Genres         []string
	ISRC           *string
	Duration       *int
}

func extractRecordingID(url string) string {
	matches := recordingURLPattern.FindStringSubmatch(url)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

func parseQuery(query string) (title, artist string) {
	parts := strings.SplitN(query, " - ", 2)
	if len(parts) == 2 {
		artist = strings.TrimSpace(parts[0])
		title = strings.TrimSpace(parts[1])
	} else {
		title = strings.TrimSpace(query)
	}
	return
}

func recordingToMetadata(rec *musicbrainz.RecordingDetail) *sdk.SongMetadata {
	metadata := &sdk.SongMetadata{
		Title: rec.Title,
	}

	if len(rec.ArtistCredit) > 0 {
		artistName := rec.ArtistCredit[0].Artist.Name
		metadata.Artist = &artistName
	}

	if len(rec.Releases) > 0 {
		albumTitle := rec.Releases[0].Title
		metadata.Album = &albumTitle
	}

	if rec.Length != nil {
		durationSec := *rec.Length / 1000
		metadata.Duration = &durationSec
	}

	return metadata
}

func searchRecordingToMetadata(rec *musicbrainz.Recording) *sdk.SongMetadata {
	metadata := &sdk.SongMetadata{
		Title: rec.Title,
	}

	if len(rec.ArtistCredit) > 0 {
		artistName := rec.ArtistCredit[0].Artist.Name
		metadata.Artist = &artistName
	}

	if len(rec.Releases) > 0 {
		albumTitle := rec.Releases[0].Title
		metadata.Album = &albumTitle
	}

	return metadata
}
