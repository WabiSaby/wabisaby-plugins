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

package coordinator

import (
	"context"
	"fmt"

	nodepb "github.com/wabisaby/wabisaby-protos-go/go/node"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client is a client for the WabiSaby Node Coordinator gRPC service.
type Client struct {
	conn   *grpc.ClientConn
	client nodepb.NodeCoordinatorClient
}

// NewClient creates a new coordinator client.
func NewClient(address string) (*Client, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to coordinator: %w", err)
	}

	return &Client{
		conn:   conn,
		client: nodepb.NewNodeCoordinatorClient(conn),
	}, nil
}

// Close closes the connection to the coordinator.
func (c *Client) Close() error {
	return c.conn.Close()
}

// StoreContent notifies the coordinator about new content.
func (c *Client) StoreContent(ctx context.Context, req *nodepb.IndexContentRequest) (*nodepb.IndexContentResponse, error) {
	return c.client.IndexContent(ctx, req)
}
