<p align="center">
  <img src="https://imgur.com/Y59vox1.png" width="400"/>
  <br/>
  <b>-- a production of [<a href="https://github.com/0xfe">0xfe industries</a>] --</b>
</p

Lumen is a batteries-included commandline client for the Stellar blockchain. It's designed to
be robust, easy-to-use, and embeddable in shell scripts.

Lumen is based on MicroStellar, and is designed for the @qubit-sh Microbanking platform.

## QuickStart

### Installation

```bash
go get github.com/0xfe/lumen
```

### Usage

#### Make a payment and check your balance

```bash
# Pay 4 lumens from SCS... to GAU...
lumen pay SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM 4

# Check your balance
lumen balance GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM
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
lumen pay bob mo 5

# What's Mo's address?
lumen account get mo address

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
lumen pay mo bob 5 USD-chase
```

#### Generate keys and create trust lines

```bash

# Generate a new keypair (address and seed) with the alias mo
lumen account new kelly

# Fund it with some XLM to create a valid account
lumen pay mo kelly 10 --fund

# Create a trustline for kelly to Citibank's USD, then pay her
lumen trust kelly USD-citi
lumen pay mo kelly 5 USD-citi --memotext "here's five bucks"
```

#### Stream the ledger to watch for payments

```bash
# Watch for payments to and from kelly. This runs forever and emits payment details everytime
# there's a payment
lumen watch payments kelly

# Watch for payments from a known starting point
lumen watch payments kelly --cursor

# Watch for payments across multiple addresses
lumen watch payments kelly mo bob
```

#### Multisig accounts

```bash
# Create aliases for mary, sharon, and bob
lumen account set mary GDRTX6RFQULJMB4RXDNNAUNIZPLLINISMNXV4WQVXQFQBHAMPMBEWLFT SDXWOG4ZNW5RLTROHPFKCDSKKEFVKZYI4SLZIO6TXM6FJ7CKUCO5NWYB
lumen account set sharon GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM SCSJQEK352QDSXZWELWC2NKKQL6BAUKE7EVS56CKKRDQGY6KCYLRWCVQ
lumen account set bill GBJDIMENGOKR49V63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVOFI53G SBLPAE53C6JXKX6CK4UN7DIXMD4EXGA4QL6NB63YHGZRTG6NPXAPWQTC

# Add sharon as a signer on mary's account with a weight of 1
lumen signer add mary sharon 1

# Add bill as a signer too
lumen signer add mary bill 1

# Set mary's low, medium, and high thresholds to require a minimum total weight of 2
# for all transactions
lumen thresholds set mary 2 2 2

# Now mary needs atleast two signatures (including hers) to make payments
lumen pay mary mo 4 --signers mary,bill
lumen pay mary bob 10 USD --signers sharon,bill
```

#### Advanced

```bash

# Kill mary's master key
lumen setmasterweight mary 0

# Set mary's home domain
lumen options set homedomain mary qubit.sh

# Whoops! we killed her master key. Fortunately she has other signers on her account.
lumen options set homedomain mary qubit.sh --signers bill,sharon

# Who are mo's signers?
lumen options get mo signers
```