package awssecurityhub

import (
	"context"

	"github.com/theopenlane/core/v2/internal/integrations/types"
)

// assumeRoleClient returns the builder of the AWS client under STS assume-role
func assumeRoleClient(opCfg Config) func(context.Context, types.ConnectionRequest[AssumeRoleCredentialSchema]) (Client, error) {
	return func(ctx context.Context, req types.ConnectionRequest[AssumeRoleCredentialSchema]) (Client, error) {
		cred := req.Credential

		switch {
		case cred.RoleARN == "":
			return Client{}, ErrRoleARNMissing
		case cred.HomeRegion == "":
			return Client{}, ErrRegionMissing
		}

		cfg, err := buildAWSConfig(ctx, cred, opCfg)
		if err != nil {
			return Client{}, err
		}

		return Client{
			Config: cfg,
			Scope: CollectionScope{
				AccountID:     arnAccountID(cred.RoleARN),
				AccountScope:  cred.AccountScope,
				AccountIDs:    cred.AccountIDs,
				LinkedRegions: cred.LinkedRegions,
			},
		}, nil
	}
}

// staticClient builds the AWS client from static IAM credentials
func staticClient(ctx context.Context, req types.ConnectionRequest[ServiceAccountCredentialSchema]) (Client, error) {
	cfg, err := buildAWSConfigFromStaticCreds(ctx, req.Credential)
	if err != nil {
		return Client{}, err
	}

	return Client{
		Config: cfg,
		Scope:  CollectionScope{AccountScope: AccountScopeAll},
	}, nil
}
