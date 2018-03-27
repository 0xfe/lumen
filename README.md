<p align="center">
  <img src="https://imgur.com/Y59vox1.png" width="400"/>
  <br/>
</p>

Lumen is a batteries-included commandline client for the Stellar blockchain. It's designed to
be easy-to-use, robust, and embeddable (in both shell scripts and other Go applications.) [You can download Lumen here.](https://github.com/0xfe/lumen/releases)

Learn all about Lumen in the open-source Stellar e-book, [Hacking Stellar](http://github.com/0xfe/hacking-stellar).

<a href="https://travis-ci.org/0xfe/lumen"><img src="https://travis-ci.org/0xfe/lumen.svg?branch=master"/></a>

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

# Make cross currency payments. Bob pays Mary 20 USD with EUR (spending no more than 10 EUR.)
$ lumen pay 20 USD --from bob --to mary --with EUR --max 10
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
* Decode, sign, and submit arbitrary transactions
  ```bash
  # Output a base64 transaction without signing or submitting it to the network.
  lumen pay 5 USD --from mary --to bob --nosign --nosubmit >payment.txt

  # Decode a base64-encoded transaction into a JSON string.
  lumen tx decode $(cat payment.txt) --pretty
  # Output: JSON transaction string

  # Add a signature to an encoded transaction.
  lumen tx sign $(cat payment.txt) --signers mary,pizzafund >payment.signed.txt

  # Submit a base64-encoded transaction to the network.
  lumen tx submit $(cat payment.signed.txt)
  # Output: horizon response
  ```
* Use federated addresses directly in your transactions
  ```bash
  lumen pay 10 --from kelly --to mo*qubit.sh
  lumen balance mo*qubit.sh

  # Works for assets too
  lumen balance bob USD:issuer*citibank.com
  ```
* Trade assets on the DEX and make path payments with automatic path finding
  ```bash
  # List all DEX trades between USD and XLM
  lumen dex orderbook USD native

  # Sell 10 USD for EUR at 2 EUR/USD (i.e, buy 5 EUR for 10 USD)
  lumen dex trade bob --sell USD --buy EUR --amount 10 --price 2

  # List bobs trade offers
  lumen dex list bob --limit 5

  # Cross-asset payments (path payments) via the DEX
  lumen pay 20 USD --from bob --to mary --with native --max 10 --path EUR,INR

  # If you don't speficy --path, Lumen finds a path for you!
  lumen pay 20 USD --from bob --to mary --with EUR --max 10
  ```
* Embed Lumen into your own Go applications
  ```go
  import (
    "github.com/0xfe/lumen/cli"
  )

  func main() {
    lumen := cli.NewCLI().Embeddable()
    lumen.RunCommand("pay 10 --from mo --to bob")
  }
  ```

* Supports almost all [MicroStellar](https://github.com/0xfe/microstellar) operations (multisig, streaming, etc.)

Lumen is based on [MicroStellar](https://github.com/0xfe/microstellar), and is designed for the @qubit-sh Microbanking platform.

## QuickStart

### Download and install

Download a release from the [Releases](https://github.com/0xfe/lumen/releases) tab and copy it to your path.

```bash
sudo cp ~/Downloads/lumen.macos /usr/bin/lumen
```

To get the Go library via `go get`:

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

# Send a payment with verbose logging
$ lumen pay 10 USD --from mo --to mary -v
# DEBU[0000] LUMEN_ENV not set                             type=setup
# DEBU[0000] using storage driver file with /Users/mo/.lumen-data.json  type=setup
# DEBU[0000] getting global:ns                             method=GetGlobalVar type=cli
# DEBU[0000] got val: default (expires: false, expires_on: 2018-03-21 08:00:43.557659 -0400 EDT)  key="global:ns" method=get type=filestore
# DEBU[0000] getting default:vars:config:network           method=GetVar type=cli
# DEBU[0000] getting default:asset:USD:code                method=GetVar type=cli
# DEBU[0000] got val: USD (expires: false, expires_on: 2018-03-21 08:02:49.189346 -0400 EDT)  key="default:asset:USD:code" method=get type=filestore
# DEBU[0000] getting default:asset:USD:issuer              method=GetVar type=cli
# DEBU[0000] got val: GAGUZYRM2G7235EM3C3WY33UHTWORNR7MX2J3OT4F3D46HAFSUEA63LL (expires: false, expires_on: 2018-03-21 08:02:49.188011 -0400 EDT)  key="default:asset:USD:issuer" method=get type=filestore
# DEBU[0000] got asset: &{Code:USD Issuer:GAGUZYRM2G7235EM3C3WY33UHTWORNR7MX2J3OT4F3D46HAFSUEA63LL Type:credit_alphanum4}
# DEBU[0000] getting default:account:mo:seed               method=GetVar type=cli
# DEBU[0000] got val: SAZW22BHCNGEIQNSTWR5OYG3HC44CTPXV4342OEQJDJZWCPHYUQLZCC7 (expires: false, expires_on: 2018-03-06 09:14:24.691781 -0500 EST)  key="default:account:mo:seed" method=get type=filestore
# DEBU[0000] getting default:account:mary:address          method=GetVar type=cli
# DEBU[0000] got val: GD6JJSOKWI7U2YDCMZ3YGPKNOP6W3D7K34HWLC6WHD32CKJJVALV7OBK (expires: false, expires_on: 2018-03-21 08:01:13.923333 -0400 EDT)  key="default:account:mary:address" method=get type=filestore
# DEBU[0000] paying 10 USD/GAGUZYRM2G7235EM3C3WY33UHTWORNR7MX2J3OT4F3D46HAFSUEA63LL from SAZW22BHCNGEIQNSTWR5OYG3HC44CTPXV4342OEQJDJZWCPHYUQLZCC7 to GD6JJSOKWI7U2YDCMZ3YGPKNOP6W3D7K34HWLC6WHD32CKJJVALV7OBK, opts: &{ctx:<nil> handlers:map[] hasFee:false fee:0 hasTimeBounds:false timeBounds:0 memoType:0 memoText: memoID:0 skipSignatures:false signerSeeds:[] hasCursor:false cursor: hasLimit:false limit:0 sortDescending:false passiveOffer:false sourceAddress: sendAsset:<nil> maxAmount: path:[] isMultiOp:false multiOpSource:}  cmd=pay
# DEBU[0000] signing transaction, seq: 33366067619299340   lib=microstellar method=Tx.Sign
# DEBU[0000] signed transaction, payload: AAAAAPGR63kaYI062wyHd+LARbBzZOCK9pDleNq8UkGhV4sZAAAAZAB2ikMAAAAMAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAA/JTJyrI/TWBiZneDPU1z/W2P6t8PZYvWOPehKSmoF18AAAABVVNEAAAAAAANTOIs0b+t9IzYt2xvdDzs6LY/ZfSdunwux88cBZUIDwAAAAAF9eEAAAAAAAAAAAGhV4sZAAAAQIJduVNXFgBu3/OD6uLLJJlkZD4i8JoHHorCxKi0L0LnbnVvsl2pVuazburcSH43N6AYPHI9kD/M6B03kZaz4gg=  lib=microstellar method=Tx.Sign
# DEBU[0000] submitting transaction to network test        lib=microstellar method=Tx.Submit
# DEBU[0001] transaction submitted to ledger 8026171 with hash abbac2c2906342dff927c7a88075487418c787bc4550fea6353dfc2c2faa75b2  lib=microstellar method=Tx.Submit
```

#### Create aliases

It's a pain in the butt to keep typing in addresses and seeds. Lumen lets you create aliases for your
accounts and assets to make your life easier. These aliases are stored in `$HOME/.lumen-data.json`.

```bash
# Make an alias for Mo. Lumen knows it's an address and not a seed from the string format.
lumen account set mo GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# Make an alias for Bob. Again, Lumen knows it's a seed and not an address.
lumen account set bob SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ

# Generate a new random keypair (address and seed) with the alias mo
lumen account new mary

# What's Mary's address?
lumen account address mary

# Use --fund to fund it with some XLM to create a valid account. This is required
# for all new accounts before you can transact on them.
lumen pay 1 --from mo --to mary --fund

# Bob pays Mo 5 XLM
lumen pay 5 --from bob --to mo

# Lookup federated addresses
lumen account address mo*qubit.sh
lumin account set mo mo*qubit.sh

# Check Mo's balance (this shows the balance of mo*qubit.sh)
lumen balance mo
```

#### Work with credit assets

```bash
# Look up bob's USD balance (from specified issuer)
lumen balance bob USD:GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# Create an alias for a new asset type. The asset code is derived from the alias (USD).
lumen asset set USD GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM

# If you want to specify an asset code different from the alias.
lumen asset set USD-citi GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM --asset-code USD
lumen asset set USD-chase GBGFCNBK5ITK5PTCXDTB3XPDYY4UHZAWMX77YXEEV5QPANLELZLC7MXA --asset-code USD

# Check bob's USD balance
lumen balance bob USD-chase

# Create a trustline for kelly to Citibank's USD, then pay her
lumen trust create kelly USD-citi
lumen pay 5 USD-citi --from mo --to kelly --memotext "here's five bucks"

# Use federated asset names
lumen pay 5 USD:issuer*chase.com --from mo --to kelly --memotext "here's five bucks"

# Require issuer to authorize all new trustlines (and make them revocable)
lumen flags issuer auth_required auth_revocable

# Create a new trustline and authorize it
lumen trust create kelly USD-citi
lumen trust allow kelly USD-citi --signers citibank

# Revoke a trust line (or freeze an asset in an account)
lumen trust allow kelly USD-citi --revoke --signers citibank
```

#### Stream the ledger

```bash
# Watch for payments to and from kelly. This runs forever and emits payment details everytime
# there's a payment
lumen watch payments kelly

# Stream payments all the way from when the account was created
lumen watch payments kelly --cursor start

# Stream all transactions from kelly
lumen watch transactions kelly

# Stream all ledger updates in Stellar
lumen watch ledger
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
lumen signer add sharon --to mary 1

# Add bill as a signer too
lumen signer add bill --to mary 1

# Set mary's low, medium, and high thresholds to require a minimum total weight of 2
# for all transactions
lumen signer thresholds mary 2 2 2

# Now mary needs atleast two signatures (including hers) to make payments
lumen pay 4 --from mary --to mo --signers mary,bill
lumen pay 10 USD --from mary --to bob --signers sharon,bill

# Remove bill as a signer
lumen signer remove bill --from mary --signers mary,bill
```

#### Advanced features

```sh
# Cross-asset path payments. Deposit 10 INR into Mary's account using USD from
# Bob's USD account, spending no more than 3 USD
lumen pay 10 INR --to mary --from bob --with USD --max 3

# Attach data fields to an account
lumen data bob mydata "the fresh prince"
lumen data bob otherdata "more data"

# Lookup the key "mydata" in bob's account
lumen data bob mydata
# output: the fresh prince

# Delete data key mydata
lumen data bob mydata --clear

# Display a base64 transaction signed by mary without submitting it to the network
lumen pay 5 USD --from mary --to bob --nosubmit
# Output: base64-encoded transaction

# Display an unsigned base64 transaction without submitting it to the network
lumen pay 5 USD --from mary --to bob --nosign --nosubmit
# Output: base64-encoded transaction

# Decode a base64-encoded transaction
lumen tx decode AAAAALiDDp5...

# Add a signature to an encoded transaction
lumen tx sign AAAAALiDDp5... --signers mary,pizzafund
# Output: signed base64 transaction

# Submit a base64-encoded transaction to the network
lumen tx submit AAAAALiDDp5...
# Output: horizon response

# Get detailed account information in JSON
lumen info bob

# Change bob's account flags
lumen flags bob auth_revocables

# Disable Bob's master key (by setting it's weight to 0)
lumen signer masterweight bob 0

# Create a time bound transaction only valid between given UTC timestamps
# Submit it later with: lumen tx submit "base64-encoded transaction string"
lumen pay 5 USD --from escrow --to bob --mintime '2017-06-06 12:00:00' --maxtime '2017-05-05 12:00:00' --nosubmit
# Output: base64-encoded transaction string
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

## Hacking on Lumen

### Contribution Guidelines

* We're managing dependencies with [dep](https://github.com/golang/dep).
  * Add a new dependency with `dep ensure -add ...`
* If you're adding a new feature:
  * Add unit tests
  * Add godoc comments
  * If necessary, update the integration test in `lumen_test.go`

**You can also support this project by sending lumens to GDEVC4BOVFMB46UHGJ6NKEBCQVY5WI56GOBWPG3QKS4QV4TKDLPE6AH6.**

### Environment Setup

This package uses [dep](https://github.com/golang/dep) to manage dependencies. Before
hacking on this package, install all dependencies and build.

```
dep ensure
make
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

### Releasing

We're managing releases with [gothub](https://github.com/itchio/gothub).

```bash
go get github.com/itchio/gothub

# find token in ~/.git-credentials
export GITHUB_TOKEN=...

# test, commit, and build binaries
git test ./...
git commit -a
git push
make

# push new release
./release.sh vA.B.C
```

## MIT License

Copyright Mohit Muthanna Cheppudira 2018 <mohit@muthanna.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
