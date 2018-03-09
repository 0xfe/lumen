<p align="center">
  <img src="https://imgur.com/Y59vox1.png" width="400"/>
  <br/>
</p>

Lumen is a batteries-included commandline client for the Stellar blockchain. It's designed to
be easy-to-use, robust, and embeddable (in both shell scripts and other Go applications.)

```bash
# Switch to the test network
$ lumen set config:network test

# Create two new accounts: bob and mary
$ lumen account new mary
# GDGO2Z2556NQ2JXFHQH2CUIF3E54KTHZIDA7EGUZSG4TJXPV6YZ4MEI4 SBJ24KK6HWOF44MWFH3J7VPAGOMFOFO7O23YW6H37MYZU6F44JBYKLMU

$ lumen account new bob
# GC7BE5UFOO3BFHMNS6G66ANMLP4K22KIIIZAY6L6X67RWQA5GVHNBVNG SAHBBWKIX3CBXXCY6R6NNFCJMB33GNOTFVMUZ3WGVCWL5WXHU7HOOJLI

# Fund Mary via friendbot
$ lumen friendbot mary

# Friend bob via mary (we use the --fund flag to specify that this is a new account)
$ lumen pay 1000 --from mary --to bob --fund

# Pay Mary back
$ lumen pay 5 --from bob --to mary --memotext 'thanks for the fish'

# Check Bob's balance
$ lumen balance bob
# 994.99990
```

#### Some notable features:

* Use account and asset aliases instead of addresses directly in your commands.
  ```bash
  lumen asset set USD-chase SBJ24KK6HWOF44MWFH3J7VPAGOMFOFO7O23YW6H37MYZU6F44JBYKLMU --code USD
  lumen pay 10 USD-chase --from kelly --to bob

  lumen account set issuer_tdbank SD3EEU6KJYE6DSEQQYXEXKNOAKZWPNT5OROUGZ5OT2CPFGSLD2JF4ZGY
  lumen asset set CAD issuer_tdbank
  lumen pay 10 CAD --from kelly --to bob
  ```
* Segregate your aliases and configuration using namespaces.
  ```bash
  # Create and switch to new namespace
  lumen ns manhattan_project

  # This namespace always operates on the public network
  lumen set config:network public
  lumen pay 100000 USD --from president --to terrorist
  ```
* Share addresses and encrypted seeds with other users via Redis.
  ```bash
  export LUMEN_STORE="redis,localhost:3400"

  # Aliases are loaded and saved from redis
  lumen account new ally
  lumen friendbot ally
  ```
* Embed Lumen into your own Go applications
  ```go
  import (
    "fmt"
    "io/ioutil"
    "log"
    "os"
    "strconv"
    "strings"
    "testing"

    "github.com/0xfe/lumen/cli"
    "github.com/sirupsen/logrus"
  )

  func main() {
    lumen := cli.NewCLI().Embeddable()
    lumen.RunCommand("pay 10 --from mo --to bob")
  }
  ```

* Supports almost all [MicroStellar](https://github.com/0xfe/microstellar) operations (multisig, streaming, etc.)

Lumen is based on [MicroStellar](https://github.com/0xfe/microstellar), and is designed for the @qubit-sh Microbanking platform.

## QuickStart

### Installation

```bash
go get github.com/0xfe/lumen
```

### Usage

#### Make a payment and check your balance

```bash
# Pay 4 lumens from SCS... to GAU...
lumen pay 4 --from SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ --to GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# Check your balance
lumen balance GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM
```

Lumen defaults to the test network for all operations. To use the public network, use the `--network public` flag,
or call `lumen set config:network public`.

```bash
# Try it on the public network
lumen balance GBDZI7NMPUMAIWMHUZWJK5EOGAQZTL3GKWMKX2LSQNYDE42NYY7SLJRB --network public

# Or...
lumen set config:network public
lumen balance GBDZI7NMPUMAIWMHUZWJK5EOGAQZTL3GKWMKX2LSQNYDE42NYY7SLJRB
```

#### Create aliases

It's a pain in the butt to keep typing in addresses and seeds. Lumen lets you create aliases for your
accounts and assets to make your life easier. These aliases are stored in `$HOME/.lumen-data.json`.

```bash
# Make an alias for Mo. Lumen knows it's an address and not a seed from the string format.
lumen account set mo GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# Make an alias for Bob. Again, Lumen knows it's a seed and not an address.
lumen account set bob SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ

# Bob pays Mo 5 XLM
lumen pay 5 --from bob --to mo

# What's Mo's address?
lumen account address mo

# What's Mo's seed?
lumen account seed mo

# Check Mo's balance
lumen balance mo
```

#### Work with assets

```bash
# Create an alias for a new asset type. The asset code is derived from the alias (USD).
lumen asset set USD GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# If you want to specify an asset code different from the alias.
lumen asset set USD-citi GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM --asset-code USD
lumen asset set USD-chase GBGFCNBK5ITK5PTCXDTB3XPDYY4UHZAWMX77YXEEV5QPANLELZLC7MXA --asset-code USD

# Pay with the asset
lumen pay --from mo --to bob 5 USD-chase

# Check bob's USD balance
lumen balance bob USD-chase
```

#### Generate keys and create trust lines

```bash
# Generate a new random keypair (address and seed) with the alias mo
lumen account new kelly

# What's kelly's address?
lumen account address kelly

# Use --fund to fund it with some XLM to create a valid account. This is required
# for all new accounts before you can transact on them.
lumen pay 1 --from mo --to kelly --fund

# Create a trustline for kelly to Citibank's USD, then pay her
lumen trust create kelly USD-citi
lumen pay 5 USD-citi --from mo --to kelly --memotext "here's five bucks"
```

#### Working with namespaces
You can use namespaces to work on multiple projects at the same time. The default namespace is called `default`.

```bash
# Lookup the current namespace
lumen ns

# Change to namespace prod (creates a new namespace, if necessary)
lumen ns prod

# Associate this namespace with the public horizon servers
lumen set config:network public

# Setup a new account in this namespace
lumen account set corp GBPQN4UDRR7BVTSSBQFUEQ5UIJS5EJ4LRXP4TJZF5Q6IDY6OBCB6UPZR SAQBE63WGYQACOXGDSW4JEG6IXGRRRCGFW6ET3F5T4STKXSQSKRRAH2I

# Switch back to default namespace
lumen ns default

# The corp account should not exist
lumen account address corp
```

#### Stream the ledger to watch for payments

```bash
# Watch for payments to and from kelly. This runs forever and emits payment details everytime
# there's a payment
lumen watch payments kelly

# Watch for payments from a known starting point
lumen watch payments kelly --cursor ABCD

# Watch for payments across multiple addresses
lumen watch payments kelly mo bob
```

#### Multisig accounts

```bash
# Lets switch to a new namespace for this
lumen ns multisig

# Create aliases for mary, sharon, and bob
lumen account set mary GDRTX6RFQULJMB4RXDNNAUNIZPLLINISMNXV4WQVXQFQBHAMPMBEWLFT SDXWOG4ZNW5RLTROHPFKCDSKKEFVKZYI4SLZIO6TXM6FJ7CKUCO5NWYB
lumen account set sharon GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ
lumen account set bill GBJDIMENGOKR49V63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVOFI53G SBLPAE53C6JXKX6CK4UN7DIXMD4EXGA4QL6NB63YHGZRTG6NPXAPWQTC

# Add sharon as a signer on mary's account with a weight of 1
lumen signer add sharon --to mary --weight 1

# Add bill as a signer too
lumen signer add bill --to mary --weight 1

# Set mary's low, medium, and high thresholds to require a minimum total weight of 2
# for all transactions
lumen thresholds set mary 2 2 2

# Now mary needs atleast two signatures (including hers) to make payments
lumen pay mary mo 4 --signers mary,bill
lumen pay mary bob 10 USD --signers sharon,bill

# Remove bill as a signer
lumen signer remove bill --from mary
```

### Configuring Lumen

Lumen looks for a configuration file called `.lumen-config.yml` in one of the following locations (in order of preference):


* The current directory: `.`
* The parent directory: `..`
* Your home directory: `$HOME/.lumen/`
* In `/etc/lumen`

Setting `$LUMEN_ENV` to `prod` or `test`, switches the config file name to `.lumen-config-prod.yml` or `.lumen-config-test.yml` respectively.

The supported configuration format:

```yaml
# Where to store lumen data.
store:
  driver: "file"  # Other options: redis, internal (memdb for testing)
  params: "/home/mo/.lumen-data.json" # If redis, then host:port

# You can also use the -v flag to enable verbose logging.
verbose: false
```

### Data storage

By default Lumen stores data in `$HOME/.lumen-data.json`. You can change the data location by (in order of preference):

* The `--store` flag. (E.g., `lumen balance mo --store ./test.json`)
* The `LUMEN_STORE` environment variable: `export LUMEN_STORE="/etc/lumen/data.json"`
* The configuration file (see above.)

### Namespaces

Namespaces are a convenience feature that allow you to work on different projects at the same time. Namespaces
are tied to a data file, so each file can contain multiple namespaces. The pointer to the current namespace
is also stored in the data file.

You can switch namespaces by (in order of preference):

* The `--ns` flag.
* The `LUMEN_NS` environment variable.

You can get the current namespace with `lumen ns`. The default namespace is `default`.

## Hacking on Lumen

### Contribution Guidelines

* We're managing dependencies with [dep](https://github.com/golang/dep).
  * Add a new dependency with `dep ensure -add ...`
* If you're adding a new feature:
  * Add unit tests
  * Add godoc comments
  * If necessary, update the integration test in `lumen_test.go`

### Environment Setup

This package uses [dep](https://github.com/golang/dep) to manage dependencies. Before
hacking on this package, install all dependencies:

```
dep ensure
```

### Run tests

Run all unit tests:

```
go test -v ./...
```

To update dependencies:

```
# Updates dependencies in vendor/ to latest tags/releases
dep ensure -update

# rinse and repeat
go test -v ./...
```

### Versioning

This package uses semantic versioning:

```
git tag v0.1.0
git push --tags
```

## MIT License

Copyright Mohit Muthanna Cheppudira 2018 <mohit@muthanna.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
