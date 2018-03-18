package microstellar

import (
	"encoding/base64"

	"github.com/stellar/go/clients/horizon"
)

// Address represents a stellar address or public key
type Address string

// Seed represents a stellar seed or private
type Seed string

// KeyPair represents a key pair for a signer on a stellar account. An account
// can have multiple signers.
type KeyPair struct {
	Seed    string // private key
	Address string // public key
}

// Balance is the balance amount of the asset in the account.
type Balance struct {
	Asset  *Asset `json:"asset"`
	Amount string `json:"amount"`
	Limit  string `json:"limit"`
}

// Signer represents a key that can sign for an account.
type Signer struct {
	PublicKey string `json:"public_key"`
	Weight    int32  `json:"weight"`
	Key       string `json:"key"`
	Type      string `json:"type"`
}

// Thresholds represent the signing thresholds on the account
type Thresholds struct {
	High   byte `json:"high"`
	Medium byte `json:"medium"`
	Low    byte `json:"low"`
}

// Flags contains the auth flags in the account
type Flags struct {
	AuthRequired  bool `json:"auth_required"`
	AuthRevocable bool `json:"auth_revocable"`
	AuthImmutable bool `json:"auth_immutable"`
}

// Account represents an account on the stellar network.
type Account struct {
	Address       string            `json:"address"`
	Balances      []Balance         `json:"balances"`
	Signers       []Signer          `json:"signers"`
	Flags         Flags             `json:"flags"`
	NativeBalance Balance           `json:"native_balance"`
	HomeDomain    string            `json:"home_domain"`
	Thresholds    Thresholds        `json:"thresholds"`
	Data          map[string]string `json:"data"`
	Sequence      string            `json:"seq"`
}

// newAccount creates a new initialized account
func newAccount() *Account {
	account := &Account{}
	account.NativeBalance = Balance{NativeAsset, "0", ""}
	account.Signers = []Signer{
		Signer{},
	}
	account.Balances = []Balance{}

	return account
}

// newAccountFromHorizon creates a new account from a Horizon JSON response.
func newAccountFromHorizon(ha horizon.Account) *Account {
	account := newAccount()

	account.Address = ha.HistoryAccount.AccountID
	account.HomeDomain = ha.HomeDomain
	account.Sequence = ha.Sequence

	for _, b := range ha.Balances {
		if b.Asset.Type == string(NativeType) {
			account.NativeBalance = Balance{NativeAsset, b.Balance, ""}
			continue
		}

		balance := Balance{
			Asset:  NewAsset(b.Asset.Code, b.Asset.Issuer, AssetType(b.Asset.Type)),
			Amount: b.Balance,
			Limit:  b.Limit,
		}

		account.Balances = append(account.Balances, balance)
	}

	account.Signers = []Signer{}
	for _, s := range ha.Signers {
		signer := Signer{
			PublicKey: s.PublicKey,
			Weight:    s.Weight,
			Key:       s.Key,
			Type:      s.Type,
		}
		account.Signers = append(account.Signers, signer)
	}

	account.Thresholds.High = ha.Thresholds.HighThreshold
	account.Thresholds.Medium = ha.Thresholds.MedThreshold
	account.Thresholds.Low = ha.Thresholds.LowThreshold

	account.Flags.AuthRequired = ha.Flags.AuthRequired
	account.Flags.AuthRevocable = ha.Flags.AuthRevocable

	account.Data = map[string]string{}
	for k, v := range ha.Data {
		account.Data[k] = v
	}

	return account
}

// GetBalance returns the balance for asset in account. If no balance is
// found for the asset, returns "".
func (account *Account) GetBalance(asset *Asset) string {
	if asset.IsNative() {
		return account.GetNativeBalance()
	}

	for _, b := range account.Balances {
		if asset.Equals(*b.Asset) {
			return b.Amount
		}
	}

	return ""
}

// GetNativeBalance returns the balance of the native currency (typically lumens)
// in the account.
func (account *Account) GetNativeBalance() string {
	return account.NativeBalance.Amount
}

// GetMasterWeight returns the weight of the primary key in the account.
func (account *Account) GetMasterWeight() int32 {
	for _, a := range account.Signers {
		if a.PublicKey == account.Address {
			return a.Weight
		}
	}

	return -1
}

// GetData decodes and returns the base-64 encoded data in "key"
func (account *Account) GetData(key string) ([]byte, bool) {
	v, ok := account.Data[key]

	if ok {
		decodedVal, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, false
		}
		return decodedVal, true
	}

	return nil, false
}
