package grpc

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Client wraps GRPC client connection.
type Client struct {
	*grpc.ClientConn
}

// NewClient creates GRPC client.
func NewClient(grpcURL string, timeout int64) (*Client, error) {
	var grpcopts []grpc.DialOption
	var conn *grpc.ClientConn
	var err error

	urls := strings.Split(grpcURL, ":")
	if len(urls) > 2 {
		panic(fmt.Sprintf("incorrect grpc endpoint: %s", urls))
	}

	if urls[1] == "443" {
		grpcopts = []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		}
		conn, err = grpc.Dial(grpcURL, grpcopts...)
		if err != nil {
			return &Client{}, fmt.Errorf("failed to connect GRPC client: %s", err)
		}
	} else {
		grpcopts = []grpc.DialOption{
			grpc.WithInsecure(),
			grpc.WithBlock(),
		}
		conn, err = grpc.DialContext(context.Background(), grpcURL, grpcopts...)
		if err != nil {
			return &Client{}, fmt.Errorf("failed to connect GRPC client: %s", err)
		}
	}

	return &Client{conn}, nil
}

// IsNotFound returns not found status.
func IsNotFound(err error) bool {
	return status.Convert(err).Code() == codes.NotFound
}
