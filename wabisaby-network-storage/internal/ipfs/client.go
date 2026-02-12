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

package ipfs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Client is an IPFS HTTP API client.
type Client struct {
	apiURL     string
	httpClient *http.Client
}

// NewClient creates a new IPFS client.
func NewClient(apiURL string) *Client {
	return &Client{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// AddResponse represents a response from IPFS add operation.
type AddResponse struct {
	Name string `json:"Name"`
	Hash string `json:"Hash"`
}

// AddFile adds a single file to IPFS and returns its CID.
func (c *Client) AddFile(ctx context.Context, filePath string) (*AddResponse, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file to form: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/api/v0/add?pin=true", c.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("IPFS add failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result AddResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// StatResponse represents a response from IPFS object stat operation.
type StatResponse struct {
	CumulativeSize int64 `json:"CumulativeSize"`
}

// AddDirectory uploads a directory to IPFS and returns the results.
func (c *Client) AddDirectory(ctx context.Context, dirPath string) ([]AddResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		relPath, err := filepath.Rel(dirPath, path)
		if err != nil {
			return fmt.Errorf("get relative path: %w", err)
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", relPath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(part, file); err != nil {
			return fmt.Errorf("copy file to form: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	writer.Close()

	url := fmt.Sprintf("%s/api/v0/add?pin=true&wrap-with-directory=true", c.apiURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []AddResponse
	decoder := json.NewDecoder(resp.Body)
	for decoder.More() {
		var result AddResponse
		if err := decoder.Decode(&result); err != nil {
			break
		}
		results = append(results, result)
	}

	return results, nil
}

// Stat gets the stat information for a CID.
func (c *Client) Stat(ctx context.Context, cid string) (*StatResponse, error) {
	url := fmt.Sprintf("%s/api/v0/object/stat?arg=%s", c.apiURL, cid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
