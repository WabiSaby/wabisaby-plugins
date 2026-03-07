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

package musicbrainz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BaseURL        = "https://musicbrainz.org/ws/2"
	CoverArtURL    = "https://coverartarchive.org"
	UserAgent      = "WabiSaby/1.0 (https://wabisaby.com)"
	RateLimitDelay = 1100 * time.Millisecond
)

type Client struct {
	httpClient *http.Client
	mu         sync.Mutex
	lastReq    time.Time
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) rateLimit(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	elapsed := time.Since(c.lastReq)
	if elapsed < RateLimitDelay {
		wait := RateLimitDelay - elapsed
		select {
		case <-time.After(wait):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	c.lastReq = time.Now()
	return nil
}

type RecordingSearchResult struct {
	Recordings []Recording `json:"recordings"`
	Count      int         `json:"count"`
}

type Recording struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Score        int            `json:"score"`
	ArtistCredit []ArtistCredit `json:"artist-credit"`
	Releases     []Release      `json:"releases"`
	Tags         []Tag          `json:"tags"`
	ISRCs        []string       `json:"isrcs"`
}

type ArtistCredit struct {
	Artist Artist `json:"artist"`
}

type Artist struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	SortName string `json:"sort-name"`
}

type Release struct {
	ID         string      `json:"id"`
	Title      string      `json:"title"`
	Date       string      `json:"date"`
	ReleaseGroup *ReleaseGroup `json:"release-group"`
}

type ReleaseGroup struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Type     string `json:"primary-type"`
}

type Tag struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// SearchRecordings queries MusicBrainz for recordings matching title and artist.
func (c *Client) SearchRecordings(ctx context.Context, title, artist string) (*RecordingSearchResult, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}

	query := fmt.Sprintf("recording:%s", url.QueryEscape(title))
	if artist != "" {
		query += fmt.Sprintf(" AND artist:%s", url.QueryEscape(artist))
	}

	reqURL := fmt.Sprintf("%s/recording?query=%s&fmt=json&limit=5", BaseURL, url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("musicbrainz returned status %d", resp.StatusCode)
	}

	var result RecordingSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// RecordingDetail contains full metadata for a recording lookup.
type RecordingDetail struct {
	ID           string         `json:"id"`
	Title        string         `json:"title"`
	Length       *int           `json:"length"`
	ArtistCredit []ArtistCredit `json:"artist-credit"`
	Releases     []Release      `json:"releases"`
	Tags         []Tag          `json:"tags"`
	ISRCs        []string       `json:"isrcs"`
}

// GetRecording fetches full details for a recording by ID.
func (c *Client) GetRecording(ctx context.Context, recordingID string) (*RecordingDetail, error) {
	if err := c.rateLimit(ctx); err != nil {
		return nil, err
	}

	reqURL := fmt.Sprintf("%s/recording/%s?fmt=json&inc=artists+releases+tags+isrcs+release-groups", BaseURL, url.PathEscape(recordingID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("musicbrainz returned status %d", resp.StatusCode)
	}

	var result RecordingDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// CoverArtResponse holds the response from Cover Art Archive.
type CoverArtResponse struct {
	Images []CoverArtImage `json:"images"`
}

// CoverArtImage represents a single cover art image.
type CoverArtImage struct {
	ID         int64    `json:"id"`
	Front      bool     `json:"front"`
	Image      string   `json:"image"`
	Thumbnails Thumbnails `json:"thumbnails"`
}

// Thumbnails contains various sized thumbnail URLs.
type Thumbnails struct {
	Small string `json:"small"`
	Large string `json:"large"`
	S250  string `json:"250"`
	S500  string `json:"500"`
	S1200 string `json:"1200"`
}

// GetCoverArt fetches cover art for a release from Cover Art Archive.
// Returns the front cover URL or empty string if none found.
func (c *Client) GetCoverArt(ctx context.Context, releaseID string) (string, error) {
	reqURL := fmt.Sprintf("%s/release/%s", CoverArtURL, url.PathEscape(releaseID))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("coverart request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("coverart returned status %d", resp.StatusCode)
	}

	var result CoverArtResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	for _, img := range result.Images {
		if img.Front {
			if img.Thumbnails.S500 != "" {
				return img.Thumbnails.S500, nil
			}
			return img.Image, nil
		}
	}

	if len(result.Images) > 0 {
		if result.Images[0].Thumbnails.S500 != "" {
			return result.Images[0].Thumbnails.S500, nil
		}
		return result.Images[0].Image, nil
	}

	return "", nil
}

// ParseReleaseYear extracts the year from a release date string.
func ParseReleaseYear(date string) *int {
	if date == "" {
		return nil
	}
	parts := strings.Split(date, "-")
	if len(parts) == 0 {
		return nil
	}
	year, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil
	}
	return &year
}
