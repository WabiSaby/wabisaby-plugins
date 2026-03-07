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

package fingerprint

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// FingerprintResult contains the output from Chromaprint analysis.
type FingerprintResult struct {
	Fingerprint string
	Duration    int
}

// fpcalcJSON is the JSON output format from fpcalc.
type fpcalcJSON struct {
	Duration    float64 `json:"duration"`
	Fingerprint string  `json:"fingerprint"`
}

// Generate runs fpcalc on an audio file and returns its fingerprint.
// Requires the fpcalc binary to be installed and in PATH.
func Generate(ctx context.Context, filePath string) (*FingerprintResult, error) {
	cmd := exec.CommandContext(ctx, "fpcalc", "-json", filePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fpcalc error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("run fpcalc: %w", err)
	}

	var result fpcalcJSON
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("parse fpcalc output: %w", err)
	}

	return &FingerprintResult{
		Fingerprint: result.Fingerprint,
		Duration:    int(result.Duration),
	}, nil
}

// GenerateWithPlainOutput runs fpcalc with plain text output as fallback.
func GenerateWithPlainOutput(ctx context.Context, filePath string) (*FingerprintResult, error) {
	cmd := exec.CommandContext(ctx, "fpcalc", filePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("fpcalc error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("run fpcalc: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	result := &FingerprintResult{}

	for _, line := range lines {
		if strings.HasPrefix(line, "DURATION=") {
			durStr := strings.TrimPrefix(line, "DURATION=")
			dur, _ := strconv.ParseFloat(durStr, 64)
			result.Duration = int(dur)
		} else if strings.HasPrefix(line, "FINGERPRINT=") {
			result.Fingerprint = strings.TrimPrefix(line, "FINGERPRINT=")
		}
	}

	if result.Fingerprint == "" {
		return nil, fmt.Errorf("no fingerprint in output")
	}

	return result, nil
}

// IsAvailable checks if fpcalc is installed and accessible.
func IsAvailable() bool {
	_, err := exec.LookPath("fpcalc")
	return err == nil
}
