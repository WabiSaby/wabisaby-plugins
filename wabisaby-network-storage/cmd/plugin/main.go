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

package main

import (
	"log"

	sdk "github.com/wabisaby/wabisaby-plugin-sdk/go"
	"github.com/wabisaby/wabisaby-plugins/wabisaby-network-storage/internal/provider"
)

func main() {
	plugin := provider.NewWabiSabyStoragePlugin()

	if err := sdk.Serve(plugin); err != nil {
		log.Fatalf("failed to start plugin: %v", err)
	}
}
