package microstellar

import (
	"github.com/pkg/errors"
	"github.com/stellar/go/build"
)

// AssetType represents an asset type on the stellar network.
type AssetType string

// NativeType represents native assets (like lumens.)
const NativeType AssetType = "native"

// Credit4Type represents credit assets with 4-digit codes.
const Credit4Type AssetType = "credit_alphanum4"

// Credit12Type represents credit assets with 12-digit codes.
const Credit12Type AssetType = "credit_alphanum12"

// Asset represents a specific asset class on the stellar network. For native
// assets "Code" and "Issuer" are ignored.
type Asset struct {
	Code   string    `json:"code"`
	Issuer string    `json:"issuer"`
	Type   AssetType `json:"type"`
}

// NativeAsset is a convenience const representing a native asset.
var NativeAsset = &Asset{"XLM", "", NativeType}

// NewAsset creates a new asset with the given code, issuer, and assetType. assetType
// can be one of: NativeType, Credit4Type, or Credit12Type.
//
//   USD := microstellar.NewAsset("USD", "issuer_address", microstellar.Credit4Type)
func NewAsset(code string, issuer string, assetType AssetType) *Asset {
	if assetType == NativeType {
		code = "xlm"
	}
	return &Asset{code, issuer, assetType}
}

// Equals returns true if "this" and "that" represent the same asset class.
func (this Asset) Equals(that Asset) bool {
	// For native assets, don't compare code or issuer
	if this.Type == NativeType || that.Type == NativeType {
		return this.Type == that.Type
	}

	return (this.Code == that.Code && this.Issuer == that.Issuer && this.Type == that.Type)
}

// IsNative returns true if the asset is a native asset (e.g., lumens.)
func (asset Asset) IsNative() bool {
	return asset.Type == NativeType
}

// Validate returns error if the asset is not valid.
func (asset Asset) Validate() error {
	if asset.Type == Credit4Type && len(asset.Code) > 4 {
		return errors.Errorf("invalid: Credit4Type assets must not have more than 4 characters")
	} else if asset.Type == Credit12Type && len(asset.Code) > 12 {
		return errors.Errorf("invalid: Credit12Type assets must not have more than 12 characters")
	}

	if !asset.IsNative() && !ValidAddressOrSeed(asset.Issuer) {
		return errors.Errorf("invalid issuer: %s", asset.Issuer)
	}

	return nil
}

// ToStellarAsset returns a stellar-go Asset from this one.
func (asset Asset) ToStellarAsset() build.Asset {
	if asset.IsNative() {
		return build.NativeAsset()
	}

	return build.CreditAsset(asset.Code, asset.Issuer)
}
