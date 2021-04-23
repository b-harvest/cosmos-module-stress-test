## Overview

This program performs stress testing for the liquidity module. This will help to prepare the upcoming [Gravity DEX Testnet Competition](https://gravitydex.io/).
## Version

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Testnet Information
### genesis.json 

- https://github.com/nodebreaker0-0/gentx/blob/main/genesis.json
- chain-id is `swap-testnet-2004` 

### Token types

```bash
"coins":[
    {
        "denom": "stake",
        "amount": "50000000000000"
    },
    {
        "denom": "uakt",
        "amount": "500000000000000"
    },
    {
        "denom": "uatom",
        "amount": "500000000000000"
    },
    {
        "denom": "ubtsg",
        "amount": "50000000000000"
    },
    {
        "denom": "udvpn",
        "amount": "500000000000000"
    },
    {
        "denom": "ugcyb",
        "amount": "500000000000000"
    },
    {
        "denom": "uiris",
        "amount": "500000000000000"
    },
    {
        "denom": "uluna",
        "amount": "500000000000000"
    },
    {
        "denom": "ungm",
        "amount": "500000000000000"
    },
    {
        "denom": "uxprt",
        "amount": "5000000000000000"
    },
    {
        "denom": "uxrn",
        "amount": "500000000000000"
    },
    {
        "denom": "xrun",
        "amount": "500000000000000"
    }
]
```

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
