package grpc

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// IsNotFound returns not found status.
func IsNotFound(err error) bool {
	return status.Convert(err).Code() == codes.NotFound
}
