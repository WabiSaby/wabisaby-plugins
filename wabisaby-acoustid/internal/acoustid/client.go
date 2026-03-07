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

package acoustid

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	BaseURL   = "https://api.acoustid.org/v2"
	UserAgent = "WabiSaby/1.0 (https://wabisaby.com)"
)

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiKey:     apiKey,
	}
}

// LookupResponse represents the AcoustID API response.
type LookupResponse struct {
	Status  string   `json:"status"`
	Results []Result `json:"results"`
}

// Result represents a single AcoustID lookup result.
type Result struct {
	ID         string      `json:"id"`
	Score      float64     `json:"score"`
	Recordings []Recording `json:"recordings"`
}

// Recording represents a MusicBrainz recording from AcoustID.
type Recording struct {
	ID       string   `json:"id"`
	Title    string   `json:"title"`
	Duration int      `json:"duration"`
	Artists  []Artist `json:"artists"`
}

// Artist represents an artist from the recording.
type Artist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// LookupResult contains the best match from fingerprint lookup.
type LookupResult struct {
	RecordingID string
	Title       string
	ArtistName  string
	ArtistID    string
	Score       float64
}

// Lookup queries AcoustID with a fingerprint and duration.
// Returns the best matching MusicBrainz recording ID.
func (c *Client) Lookup(ctx context.Context, fingerprint string, duration int) (*LookupResult, error) {
	params := url.Values{}
	params.Set("client", c.apiKey)
	params.Set("fingerprint", fingerprint)
	params.Set("duration", strconv.Itoa(duration))
	params.Set("meta", "recordings")

	reqURL := fmt.Sprintf("%s/lookup?%s", BaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("acoustid request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("acoustid returned status %d", resp.StatusCode)
	}

	var lookupResp LookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lookupResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if lookupResp.Status != "ok" {
		return nil, fmt.Errorf("acoustid status: %s", lookupResp.Status)
	}

	return findBestMatch(lookupResp.Results), nil
}

func findBestMatch(results []Result) *LookupResult {
	var best *LookupResult
	var bestScore float64

	for _, result := range results {
		for _, rec := range result.Recordings {
			score := result.Score
			if score > bestScore {
				bestScore = score
				best = &LookupResult{
					RecordingID: rec.ID,
					Title:       rec.Title,
					Score:       score,
				}
				if len(rec.Artists) > 0 {
					best.ArtistName = rec.Artists[0].Name
					best.ArtistID = rec.Artists[0].ID
				}
			}
		}
	}

	return best
}
