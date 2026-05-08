//go:build examples

package campaign

import "errors"

var (
	// ErrNameRequired is returned when --name is missing on create
	ErrNameRequired = errors.New("--name is required")
	// ErrTargetsRequired is returned when neither --emails nor --targets-file is supplied
	ErrTargetsRequired = errors.New("at least one target is required via --emails or --targets-file")
	// ErrCampaignIDRequired is returned when --campaign-id is missing on launch
	ErrCampaignIDRequired = errors.New("--campaign-id is required")
)
