package wallet_test

import (
	"testing"

	"github.com/test-go/testify/require"

	"github.com/b-harvest/liquidity-stress-test/wallet"
)

func TestRecoverAccAddrFromMnemonic(t *testing.T) {
	testCases := []struct {
		mnemonic   string
		password   string
		expAccAddr string
	}{
		{
			mnemonic:   "guard cream sadness conduct invite crumble clock pudding hole grit liar hotel maid produce squeeze return argue turtle know drive eight casino maze host",
			password:   "",
			expAccAddr: "cosmos1zaavvzxez0elundtn32qnk9lkm8kmcszzsv80v",
		},
		{
			mnemonic:   "friend excite rough reopen cover wheel spoon convince island path clean monkey play snow number walnut pull lock shoot hurry dream divide concert discover",
			password:   "",
			expAccAddr: "cosmos1mzgucqnfr2l8cj5apvdpllhzt4zeuh2cshz5xu",
		},
	}

	for _, tc := range testCases {
		accAddr, _, err := wallet.RecoverAccountFromMnemonic(tc.mnemonic, tc.password)
		require.NoError(t, err)

		require.Equal(t, tc.expAccAddr, accAddr)
	}
}
