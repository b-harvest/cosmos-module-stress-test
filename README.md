## Overview

This program performs stress testing for the liquidity module. This will help to prepare the upcoming [Gravity DEX Testnet Competition](https://gravitydex.io/).
## Version

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Testnet Information
### Genesis 

- [genesis.json](https://github.com/nodebreaker0-0/gentx/blob/main/genesis.json) file 
- chain-id is `swap-testnet-2004` 

### Coin Types

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
make install
```

### Setup local testnet

Just by running `make localnet`, this will bootstrap a single local node in your local computer and 
automatically create 4 genesis accounts with enough amounts of different types of coins. If you need more or less accounts and the other types of coins, feel free to change [the script](https://github.com/b-harvest/liquidity-stress-test/blob/main/scripts/localnet.sh#L9-L13) for your own usage.

```bash
# Run a single blockchain in your local computer 
make localnet
```

### CLI Commands

`$ tester -h`
```bash
liquidity stress testing program

Usage:
  tester [command]

Available Commands:
  create-pool create liquidity pools of every pair of coins exist in the network.
  deposit     deposit new coins to every existing pools.
  help        Help about any command
  swap        swap some coins from the exisiting pools.
  withdraw    withdraw coins from every existing pools.

Flags:
  -h, --help                help for tester
      --log-format string   logging format; must be either json or text; (default "text")
      --log-level string    logging level; (default "debug")
```