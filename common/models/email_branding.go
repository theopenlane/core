package models

// EmailBranding defines optional branding overrides for email templates.
type EmailBranding struct {
	BrandName       string `json:"brandName,omitempty"`
	LogoURL         string `json:"logoURL,omitempty"`
	PrimaryColor    string `json:"primaryColor,omitempty"`
	SecondaryColor  string `json:"secondaryColor,omitempty"`
	BackgroundColor string `json:"backgroundColor,omitempty"`
	TextColor       string `json:"textColor,omitempty"`
	ButtonColor     string `json:"buttonColor,omitempty"`
	ButtonTextColor string `json:"buttonTextColor,omitempty"`
	LinkColor       string `json:"linkColor,omitempty"`
	FontFamily      string `json:"fontFamily,omitempty"`
}

// IsZero reports whether the branding struct has no overrides set.
func (b EmailBranding) IsZero() bool {
	return b.BrandName == "" &&
		b.LogoURL == "" &&
		b.PrimaryColor == "" &&
		b.SecondaryColor == "" &&
		b.BackgroundColor == "" &&
		b.TextColor == "" &&
		b.ButtonColor == "" &&
		b.ButtonTextColor == "" &&
		b.LinkColor == "" &&
		b.FontFamily == ""
}
