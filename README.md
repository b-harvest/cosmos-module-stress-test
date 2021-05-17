<p align="center">
  <a href="https://github.com/b-harvest/gravity-dex-firestation" target="_blank"><img width="140" src="https://user-images.githubusercontent.com/20435620/117280451-92261580-ae9d-11eb-8907-f72a00320b22.jpeg" alt="B-Harvest"></a>
</p>

<h1 align="center">
    Cosmos Module Stress Testing Program 🔧
</h1>

## Overview

This program performs stress testing for the Cosmos module. Support: Liquidity , IBC transfer

**Note**: Requires [Go 1.15+](https://golang.org/dl/)
## Version

- [Liquidity Module v1.2.4](https://github.com/tendermint/liquidity/tree/v1.2.4) 
- [Cosmos SDK v0.42.4](https://github.com/cosmos/cosmos-sdk/tree/v0.42.4)
- [Tendermint v0.34.9](https://github.com/tendermint/tendermint/tree/v0.34.9)

## Usage

### Configuration

This stress testing program for the liquidity module requires a configuration file, `config.toml` in current working directory. An example of configuration file is available in `example.toml` and the config source code can be found in [here](./config.config.go).
### Build

```bash
# Clone the project 
git clone https://github.com/nodebreaker0-0/cosmos-module-stress-test.git
cd cosmos-stress-test

# Build executable
make install
```

### Setup local testnet

Just by running simple command `make localnet`, it bootstraps a single local testnet in your local computer and it
automatically creates 4 genesis accounts with enough amounts of different types of coins. You can customize them in [this script](https://github.com/nodebreaker0-0/cosmos-module-stress-testv/blob/main/scripts/localnet.sh#L9-L13) for your own usage.

```bash
# Run a single blockchain in your local computer 
make localnet
```

### CLI Commands

`$ tester -h`

```bash
comos module stress testing program

Usage:
  tester [command]

Available Commands:
  create-all-pools create liquidity pools of every pair of coins exist in the network.
  deposit     deposit new coins to every existing pools.
  help        Help about any command
  swap        swap some coins from the exisiting pools.
  transfer    Transfer a fungible token through IBC.
  withdraw    withdraw coins from every existing pools.

Flags:
  -h, --help                help for tester
      --log-format string   logging format; must be either json or text; (default "text")
      --log-level string    logging level; (default "debug")
```

## Test

### localnet

```bash
# This command is useful for local testing.
tester ca

# tester deposit [pool-id] [deposit-coins] [round] [tx-num] [flags]
tester d 1 2000000uakt,2000000uatom 5 5

# tester withdraw [pool-id] [pool-coin] [round] [tx-num] [flags]
tester w 1 10pool94720F40B38D6DD93DCE184D264D4BE089EDF124A9C0658CDBED6CA18CF27752 5 5

# tester swap [pool-id] [offer-coin] [demand-coin-denom][round] [tx-num] [msg-num]
tester s 1 1000000uakt uatom 2 2 5

# tester transfer [src-port] [src-channel] [receiver] [amount] [round] [tx-num] [msg-num]
tester transfer transfer channel-0 cosmos18zh6zd2kwtekjeg0ns5xvn2x28hgj8n6gxhe8c 1stake 1 1 1
```



