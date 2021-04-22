## Overview

This program does stress test for the liquidity module. This will help prepare the upcoming Gravity DEX testnet Competition.
## Versions

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Usage

```shell
# Download all dependencies
go mod tidy

# Run a single blockchain in your local computer
make localnet

# Run program (send MsgSend transaction)
go run main.go

# transaction hash 브라우저에서 확인
http://localhost:1317/txs/{txHash}
```