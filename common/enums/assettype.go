package enums

import "io"

// AssetType is a custom type representing the various asset types.
type AssetType string

var (
	AssetTypeTechnology AssetType = "TECHNOLOGY"
	AssetTypeDomain     AssetType = "DOMAIN"
	AssetTypeDevice     AssetType = "DEVICE"
	AssetTypeTelephone  AssetType = "TELEPHONE"
	AssetTypeInvalid    AssetType = "INVALID"
)

var assetTypeValues = []AssetType{
	AssetTypeTechnology,
	AssetTypeDomain,
	AssetTypeDevice,
	AssetTypeTelephone,
}

// Values returns a slice of strings that represents all the possible values of the AssetType enum.
// Possible default values are "TECHNOLOGY", "DOMAIN", "DEVICE", and "TELEPHONE".
func (AssetType) Values() []string { return stringValues(assetTypeValues) }

// String returns the AssetType as a string
func (r AssetType) String() string { return string(r) }

// ToAssetType returns the AssetType based on string input
func ToAssetType(r string) *AssetType { return parse(r, assetTypeValues, &AssetTypeInvalid) }

// MarshalGQL implement the Marshaler interface for gqlgen
func (r AssetType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

// UnmarshalGQL implement the Unmarshaler interface for gqlgen
func (r *AssetType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
