<p align="center">
  <img src="https://imgur.com/kPRfhrH.png" width="400"/>
  <br/>
  <b>-- a production of [<a href="https://github.com/0xfe">0xfe industries</a>] --</b>
  <br/>
</p>

**MicroStellar** is an easy-to-use Go client for the [Stellar](http://stellar.org) blockchain network. The API is simple and clean, without sacrificing flexibility.

MicroStellar is intended to be robust, well tested, and well documented -- we designed it for our Microbanking platform at @qubit-sh. It's also fun to use!

To get started, follow the instructions below, or read the [API docs](https://godoc.org/github.com/0xfe/microstellar) for more.

**MIT License:** Copyright 2018 Mohit Muthanna Cheppudira

**Build Status:** <a href="https://travis-ci.org/0xfe/microstellar"><img src="https://travis-ci.org/0xfe/microstellar.svg?branch=master"/></a>

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
pair, _ := ms.CreateKeyPair()
log.Print(pair.Seed, pair.Address)

// In stellar, you can create all kinds of asset types -- dollars, houses, kittens. These
// customized assets are called credit assets.
//
// However, the native asset is always lumens (XLM). Lumens are used to pay for transactions
// on the stellar network, and are used to fund the operations of Stellar.
//
// When you first create a key pair, you need to fund it with atleast 0.5 lumens. This
// is called the "base reserve", and makes the account valid. You can only transact to
// and from accounts that maintain the base reserve.
ms.FundAccount(
  "S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // func source
  pair.Address,                                               // fund destination
  "1")                                                        // amount in lumens (XLM)

// On the test network, you can ask FriendBot to fund your account. You don't need to buy
// lumens. (If you do want to buy lumens for the test network, call me!)
microstellar.FundWithFriendBot(pair.Address)
```

#### Check balances

```go
// Now load the account details from the ledger.
account, _ := ms.LoadAccount(pair.Address)

log.Printf("Native Balance: %v XLM", account.GetNativeBalance())
```

#### Make payments

```go
// Pay someone 3 lumens.
ms.PayNative(
  "S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // from
  "GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM", // to
  "3")

// Set the memo field on a payment
ms.Pay(
  "S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // from
  "GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM", // to
  "3", microstellar.Opts().WithMemoText("thanks for the fish"))
```

#### Work with credit assets

```go
// Create a custom asset with the code "USD" issued by some trusted issuer
USD := microstellar.NewAsset(
  "USD", // asset code
  "G6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // issuer address
  Credit4Type) // asset type (Credit4Type, Credit12Type, NativeType)

// Create a trust line from an account to the asset, with a limit of 10000
ms.CreateTrustLine(
  "S4H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // source account
  USD,     // asset to trust
  "10000") // max holdings of this asset

// Make a payment in the asset
ms.Pay(
  "S6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // from
  "GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM", // to
  USD, 10, microstellar.Opts().WithMemo("funny money"))
```

#### Multisignature payments
```go
// Add two signers with weight 1 to account
ms.AddSigner(
  "S8H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // source account
  "G6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // signer address
  1) // weight

ms.AddSigner(
  "S8H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // source account
  "G9H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGB", // signer address
  1) // weight

// Set the low, medium, and high thresholds of the account. (Here we require a minimum
// total signing weight of 2 for all operations.)
ms.SetThresholds("S8H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", 2, 2, 2)

// Kill the master weight of account, so only the new signers can sign transactions
ms.SetMasterWeight("S8H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", 0,
   microstellar.Opts().WithSigner("S2H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA"))

// Make a payment (and sign with new signers). Note that the first parameter (source) here
// can be an address instead of a seed (since the seed can't sign anymore.)
ms.PayNative(
  "G6H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA", // from
  "GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM", // to
  "3", // amount
  microstellar.Opts().
    WithSigner("S1H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA").
    WithSigner("S2H4HQPE6BRZKLK3QNV6LTD5BGS7S6SZPU3PUGMJDJ26V7YRG3FRNPGA"))

```

#### Streaming

```go
// Watch for payments to address. (The fake network sends payments every 200ms.)
watcher, err := ms.WatchPayments("GCCRUJJGPYWKQWM5NLAXUCSBCJKO37VVJ74LIZ5AQUKT6KPVCPNAGC4A")

go func() {
  for p := range watcher.Ch {
    log.Printf("WatchPayments: %v -- %v %v from %v to %v\n",
      p.Type, p.Amount, p.AssetCode, p.From, p.To)
  }
  log.Printf("WatchPayments Done -- Error: %v\n", *watcher.StreamError)
}()

// Stream the ledger for about a second then stop the watcher.
time.Sleep(1 * time.Second)
watcher.CancelFunc()
```

#### Other stuff

```go
// What's their USD balance?
account, _ = ms.LoadAccount("GAUYTZ24ATLEBIV63MXMPOPQO2T6NHI6TQYEXRTFYXWYZ3JOCVO6UYUM")
log.Printf("USD Balance: %v USD", account.GetBalance(USD))

// What's their home domain?
log.Printf("Home domain: %s", account.HomeDomain)

// Who are the signers on the account?
for i, s := range account.Signers {
    log.Printf("Signer %d (weight: %v): %v", i, s.PublicKey, s.Weight)
}
```

## Documentation

* API Docs - https://godoc.org/github.com/0xfe/microstellar
* End-to-end test - https://github.com/0xfe/microstellar/blob/master/macrotest/macrotest.go

### Supported Features

* Account creation and funding
* Lookup balances, home domain, and account signers
* Payment of native and custom assets
* Add and remove trust lines
* Multisig accounts -- add/remove signers and make multisig payments.
* Watch the ledger for streaming payments

### Coming Soon

* Offer management
* Path payments

## Hacking on MicroStellar

### Contribution Guidelines

* We're managing dependencies with [dep](https://github.com/golang/dep).
  * Add a new dependency with `dep ensure -add ...`
* If you're adding a new feature:
  * Add unit tests
  * Add godoc comments
  * If necessary, update the integration test in `macrotest/`
  * If necessary, add examples and verify that they show up in godoc

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
