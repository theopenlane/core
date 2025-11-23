package azureentraid

import "errors"

var (
	// ErrNoOrganizations indicates the Microsoft Graph API returned an empty organization list
	ErrNoOrganizations = errors.New("azureentraid: graph returned zero organizations")
)
