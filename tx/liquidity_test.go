package tx_test

import (
	"os"
	"testing"

	"github.com/b-harvest/liquidity-stress-test/client"
)

var (
	c *client.Client

	rpcAddress  = "http://localhost:26657"
	grpcAddress = "localhost:9090"
)

func TestMain(m *testing.M) {
	c = client.NewClient(rpcAddress, grpcAddress)

	os.Exit(m.Run())
}

func TestFindAllPairs(t *testing.T) {
	pairs := []struct {
		pairs []string
	}{
		{
			[]string{
				"uakt",
				"uatom",
				"ubtsg",
				"udvpn",
				"ugcyb",
				"uiris",
				"uluna",
				"ungm",
				"uxprt",
				"uxrn",
				"xrun",
			},
		},
	}

	for _, p := range pairs {
		for i := 0; i < len(p.pairs)-1; i++ {
			for j := i + 1; j < len(p.pairs); j++ {
				t.Log(p.pairs[i], p.pairs[j])
			}
		}
	}
}
