## Overview

This program does stress test for the liquidity module. This will help prepare the upcoming Gravity DEX testnet Competition.
## Version

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Usage

### Build

```bash
# Clone the project
git clone https://github.com/b-harvest/liquidity-stress-test.git
cd liquidity-stress-test
make build
```

### Setup local testnet
```bash
# Run a single blockchain in your local computer
make localnet
```
