package enums

import (
	"fmt"
	"io"
	"strings"
)

type AssetType string

var (
	AssetTypeTechnology AssetType = "TECHNOLOGY"
	AssetTypeDomain     AssetType = "DOMAIN"
	AssetTypeDevice     AssetType = "DEVICE"
	AssetTypeTelephone  AssetType = "TELEPHONE"
	AssetTypeInvalid    AssetType = "INVALID"
)

func (AssetType) Values() []string {
	return []string{
		string(AssetTypeTechnology),
		string(AssetTypeDomain),
		string(AssetTypeDevice),
		string(AssetTypeTelephone),
	}
}

func (a AssetType) String() string { return string(a) }

func ToAssetType(str string) *AssetType {
	switch strings.ToUpper(str) {
	case AssetTypeTechnology.String():
		return &AssetTypeTechnology
	case AssetTypeDomain.String():
		return &AssetTypeDomain
	case AssetTypeDevice.String():
		return &AssetTypeDevice
	case AssetTypeTelephone.String():
		return &AssetTypeTelephone
	default:
		return &AssetTypeInvalid
	}
}

func (a AssetType) MarshalGQL(w io.Writer) { _, _ = w.Write([]byte(`"` + a.String() + `"`)) }

func (a *AssetType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("wrong type for AssetType, got: %T", v) //nolint:err113
	}

	*a = AssetType(str)

	return nil
}
