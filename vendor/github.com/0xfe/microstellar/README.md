<p align="center">
  <img src="https://imgur.com/kPRfhrH.png" width="400"/>
  <br/>
</p>

**MicroStellar** is an easy-to-use Go client for the [Stellar](http://stellar.org) blockchain network. The API is simple and clean, without sacrificing flexibility.

MicroStellar is intended to be robust, well tested, and well documented -- we designed it for our Microbanking platform at @qubit-sh. It's also fun to use!

To get started, follow the instructions below, or read the [API docs](https://godoc.org/github.com/0xfe/microstellar) for more.

Also see:
* [Lumen](http://github.com/0xfe/lumen), a commandline interface for Stellar, based on MicroStellar.
* [Hacking Stellar](http://github.com/0xfe/hacking-stellar), an open-source e-book on working with Stellar.

<a href="https://travis-ci.org/0xfe/microstellar"><img src="https://travis-ci.org/0xfe/microstellar.svg?branch=master"/></a>

## QuickStart

### Installation

```
go get github.com/0xfe/microstellar
```

### Usage

#### Create and fund addresses

```go
// Create a new MicroStellar client connected to the testnet. Use "public" to
// connect to public horizon servers, or "custom" to connect to your own instance.
ms := microstellar.New("test")

// Generate a new random keypair. In stellar lingo, a "seed" is a private key, and
// an "address" is a public key. (Not techincally accurate, but you get the picture.)
//
// Seed strings begin with "S" -- "S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA"
// Address strings begin with "G" -- "GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM"
bob, _ := ms.CreateKeyPair()
log.Print(bob.Seed, bob.Address)

// In stellar, you can create all kinds of asset types -- dollars, houses, kittens. These
// customized assets are called credit assets.
//
// However, the native asset is always lumens (XLM). Lumens are used to pay for transactions
// on the stellar network, and are used to fund the operations of Stellar.
//
// When you first create a key pair, you need to fund it with atleast 0.5 lumens. This
// is called the "base reserve", and makes the account valid. You can only transact to
// and from accounts that maintain the base reserve.
ms.FundAccount(kelly.Seed, bob.Address, "1")

// On the test network, you can ask FriendBot to fund your account. You don't need to buy
// lumens. (If you do want to buy lumens for the test network, call me!)
microstellar.FundWithFriendBot(bob.Address)
```

#### Make payments and check balances

Amounts in Microstellar are typically represented as strings, to protect users from accidentaly
perform floating point operations on them. The convenience methods `ParseAmount` and `ToAmountString`
to convert the strings to `int64` values that represent a `10e7` multiple of the numeric
value. E.g., `ParseAmount("2.5") == int64(25000000)`.

```go
// Pay someone 3 lumens.
ms.PayNative(kelly.Seed, bob.Address, "3")

// Set the memo field on a payment.
err := ms.Pay(kelly.Seed, bob.Address, "3",
  microstellar.Opts().WithMemoText("thanks for the fish"))

// Find out where the transaction was submitted.
if err == nil {
  fmt.Printf("Transaction submitted to ledger: %d", ms.Response().Ledger)
}

// Get kelly's balance.
account, _ := ms.LoadAccount(kelly.Address)
log.Printf("Native Balance: %v XLM", account.GetNativeBalance())
```

#### Work with credit assets

```go
// Create a custom asset with the code "USD" issued by some trusted issuer.
USD := microstellar.NewAsset("USD", chaseBank.address, Credit4Type)

// Create a trust line from an account to the asset, with a limit of 10000.
ms.CreateTrustLine(kelly.Seed, USD, "10000")

// Make a payment in the asset.
ms.Pay(kelly.Seed, mary.Address, USD, "10",
  microstellar.Opts().WithMemo("funny money"))

// Require trustlines to be authorized buy issuer.
ms.SetFlags(issuer.Seed, microstellar.FlagAuthRequired)

// Authorize a trustline after it's been created
ms.AllowTrust(issuer.Seed, mary.Address, "USD", true)
```

#### Multisignature transactions
```go
// Add Bob as a signer to Kelly's account with the key weight 1.
ms.AddSigner(kelly.Seed, bob.Address, 1)

// Add Mary as a signer to Kelly's account.
ms.AddSigner(kelly.Seed, mary.Address, 1)

// Set the low, medium, and high thresholds of KElly's account. (Here we require a minimum
// total signing weight of 2 for all operations.)
ms.SetThresholds(kelly.Seed, 2, 2, 2)

// Disable Kelly's master key, so only Bob and Mary can sign her transactions.
ms.SetMasterWeight(kelly.Seed,
   microstellar.Opts().
     WithSigner(kelly.Seed).
     WithSigner(mary.Seed))

// Make a payment (and sign with new signers). Note that the first parameter (source) here
// can be an address instead of a seed (since the seed can't sign anymore.)
ms.PayNative(kelly.Address, pizzahut.Address, USD, "20".
  microstellar.Opts().
    WithSigner(mary.Seed).
    WithSigner(kelly.Seed))

```

#### Trade assets on the Stellar Distributed Exchange (DEX)
```go
// Sell 100 USD for lumens on the DEX at 2 lumens/USD.
err := ms.CreateOffer(bob.Seed, USD, NativeAsset, "2", "100",
  Opts().MakePassive())

// No takers, update the offer.
err := ms.UpdateOffer(bob.Seed, offerID, USD, NativeAsset, "3", "150")

// Get the order book for all USD -> Lumen trades on the DEX.
orderBook, err := ms.LoadOrderBook(USD, microstellar.NativeAsset)

// Get all offers made by Bob.
offers, err := ms.LoadOffers(bob.Seed)
```

#### Make path payments with automatic path-finding
```go
// Path payments let you transparently convert currencies. Pay 5000 INR with XLM,
// going through USD and EUR. Spend no more than 40 lumens on this transaction.
err := ms.Pay(kelly.Seed, mary.Address, "5000", INR, // mary gets 5000 INR
  microstellar.Opts().
    WithAsset(XLM, "40"). // we spend no more than 40 XLM
    Through(USD, EUR))    // go through USD and EUR

// Microstellar can automatically find paths for you, if you don't know what paths
// to take beforehand.
err := ms.Pay(kelly.Seed, mary.Address, "5000", INR, // mary receives 5000 INR
  Opts().
    WithAsset(XLM, "40"). // we spend no more than 40 XLM
    FindPathFrom(kelly.Address))
```

#### Bundle multiple operations into a single transaction
```go
// Start a mult-op transaction signed by Mary
ms.Start(bob.Address,
  microstellar.Opts().
    WithMemoText("multi-op").
    WithSigner(mary.Seed))

// Set the home domain for Bob's account.
ms.SetHomeDomain(bob.Address, "qubit.sh")

// Attach arbitrary data to Bob's account.
ms.SetData(bob.Address, "foo", []byte("bar"))

// Set Bob's acount flags.
ms.SetFlags(bob.Address, microstellar.FlagAuthRequired | microstellar.FlagsAuthImmutable)

// Make some payments
ms.Pay(bob.Address, pizzaHut.Address, USD, "500")
ms.PayNative(bob.Address, mary.Address, "25")

// Sign and submit the transaction.
ms.Submit()

// Load account to see check if it worked.
account, _ := ms.LoadAccount(bob.Address)
foo, ok := account.GetData("foo")
if ok {
  fmt.Printf("Bob's data for foo: %s", string(foo))
}

fmt.Printf("Bob's home domain: %s", account.GetHomeDomain())
```

#### Time-bound transactions for smart contracts
```go
// Create a transaction valid between 1 and 8 hours from now.
ms.Start(bob.Address,
  microstellar.Opts().WithTimeBounds(time.Now().After(1*time.Hour), time.Now().After(8.time.Hour)))
ms.Pay(bob.Address, pizzaHut.Address, USD, "500")

// Get the transaction to submit later.
payload := ms.Payload()

// A few hours later...
signedPayload, _ := ms.SignTransaction(payload, bob.Seed)
ms.SubmitTransaction(signedPayload)
```

#### Streaming

```go
// Watch for payments to and from address starting from now.
watcher, err := ms.WatchPayments(bob.Address, Opts().WithCursor("now"))

go func() {
  for p := range watcher.Ch {
    log.Printf("Saw payment on Bob's address: %+v", p)
  }
}()

// Stream the ledger for about a second then stop the watcher.
time.Sleep(1 * time.Second)
watcher.Done()

// Watch for transactions from address.
watcher, err := ms.WatchTransactions(kelly.Address, Opts().WithCursor("now"))

// Get the firehose of ledger updates.
watcher, err := ms.WatchLedgers(Opts().WithCursor("now"))
```

## Documentation

* API Docs - https://godoc.org/github.com/0xfe/microstellar
* End-to-end test - https://github.com/0xfe/microstellar/blob/master/macrotest/macrotest.go

## Hacking on MicroStellar

### Contribution Guidelines

* We're managing dependencies with [dep](https://github.com/golang/dep).
  * Add a new dependency with `dep ensure -add ...`
* If you're adding a new feature:
  * Add unit tests
  * Add godoc comments
  * If necessary, update the integration test in `macrotest/`
  * If necessary, add examples and verify that they show up in godoc

**You can also support this project by sending lumens to GDEVC4BOVFMB46UHGJ6NKEBCQVY5WI56GOBWPG3QKS4QV4TKDLPE6AH6.**

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

Run end-to-end integration test:

```
go run -v macrotest/macrotest.go
```

Test documentation with:

```
godoc -v -http=:6060
```

Then: http://localhost:6060/pkg/github.com/0xfe/microstellar/

### Updating dependencies

```
# Updates dependencies in vendor/ to latest tags/releases
dep ensure -update

# rinse and repeat
go test -v ./...
```

### Versioning

This package uses semver versioning:

```
git tag v0.1.0
git push --tags
```

## MIT License

Copyright Mohit Muthanna Cheppudira 2018 <mohit@muthanna.com>

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
