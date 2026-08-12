package enums

import "io"

type AudienceType string

var (
	AudienceTypeManual  AudienceType = "MANUAL"
	AudienceTypeDynamic AudienceType = "DYNAMIC"
	AudienceTypeInvalid AudienceType = "INVALID"
)

var audienceTypeValues = []AudienceType{
	AudienceTypeManual,
	AudienceTypeDynamic,
}

func (AudienceType) Values() []string { return stringValues(audienceTypeValues) }

func (r AudienceType) String() string { return string(r) }

func ToAudienceType(r string) *AudienceType {
	return parse(r, audienceTypeValues, &AudienceTypeInvalid)
}

func (r AudienceType) MarshalGQL(w io.Writer) { marshalGQL(r, w) }

func (r *AudienceType) UnmarshalGQL(v any) error { return unmarshalGQL(r, v) }
