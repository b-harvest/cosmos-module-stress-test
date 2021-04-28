package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client wraps GRPC client connection.
type Client struct {
	*grpc.ClientConn
}

// NewClient creates GRPC client.
func NewClient(grpcURL string, timeout int64) (*Client, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client, err := grpc.DialContext(ctx, grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return &Client{}, fmt.Errorf("failed to connect GRPC client: %s", err)
	}

	return &Client{client}, nil
}

// IsNotFound returns not found status.
func IsNotFound(err error) bool {
	return status.Convert(err).Code() == codes.NotFound
}
