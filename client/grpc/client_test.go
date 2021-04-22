package grpc_test

import (
	"os"
	"testing"

	"github.com/b-harvest/liquidity-stress-test/client/grpc"
	"github.com/b-harvest/liquidity-stress-test/codec"
)

var (
	c *grpc.Client

	grpcURL = "localhost:9090"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c = grpc.NewClient(grpcURL, 5)

	os.Exit(m.Run())
}
