package grpc

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Client wraps GRPC client connection.
type Client struct {
	*grpc.ClientConn
}

// NewClient creates GRPC client.
func NewClient(grpcURL string, timeout int64) *Client {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := grpc.DialContext(ctx, grpcURL, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		panic(fmt.Errorf("failed to connect GRPC client: %s", err))
	}

	return &Client{client}
}

// IsNotFound returns not found status.
func IsNotFound(err error) bool {
	return status.Convert(err).Code() == codes.NotFound
}
