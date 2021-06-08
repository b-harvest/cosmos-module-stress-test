package grpc_test

import (
	"os"
	"testing"

	"github.com/b-harvest/cosmos-module-stress-test/client/grpc"
	"github.com/b-harvest/cosmos-module-stress-test/codec"
)

var (
	c *grpc.Client

	grpcAddress = "localhost:9090"
)

func TestMain(m *testing.M) {
	codec.SetCodec()

	c, _ = grpc.NewClient(grpcAddress, 5)

	os.Exit(m.Run())
}
