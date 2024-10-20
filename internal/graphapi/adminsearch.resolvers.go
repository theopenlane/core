package graphapi

// THIS CODE IS REGENERATED BY github.com/theopenlane/gqlgen-plugins. DO NOT EDIT.

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
)

// Search is the resolver for the search field.
func (r *queryResolver) AdminSearch(ctx context.Context, query string) (*SearchResultConnection, error) {
	if len(query) < 3 {
		return nil, ErrSearchQueryTooShort
	}

	var (
		errors                        []error
		apitokenResults               []*generated.APIToken
		contactResults                []*generated.Contact
		documentdataResults           []*generated.DocumentData
		entitlementResults            []*generated.Entitlement
		entitlementplanResults        []*generated.EntitlementPlan
		entitlementplanfeatureResults []*generated.EntitlementPlanFeature
		entityResults                 []*generated.Entity
		entitytypeResults             []*generated.EntityType
		eventResults                  []*generated.Event
		featureResults                []*generated.Feature
		fileResults                   []*generated.File
		groupResults                  []*generated.Group
		groupsettingResults           []*generated.GroupSetting
		integrationResults            []*generated.Integration
		oauthproviderResults          []*generated.OauthProvider
		ohauthtootokenResults         []*generated.OhAuthTooToken
		organizationResults           []*generated.Organization
		organizationsettingResults    []*generated.OrganizationSetting
		personalaccesstokenResults    []*generated.PersonalAccessToken
		subscriberResults             []*generated.Subscriber
		tfasettingResults             []*generated.TFASetting
		templateResults               []*generated.Template
		userResults                   []*generated.User
		usersettingResults            []*generated.UserSetting
		webhookResults                []*generated.Webhook
	)

	r.withPool().SubmitMultipleAndWait([]func(){
		func() {
			var err error
			apitokenResults, err = searchAPITokens(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			contactResults, err = searchContacts(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			documentdataResults, err = searchDocumentData(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entitlementResults, err = searchEntitlements(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entitlementplanResults, err = searchEntitlementPlans(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entitlementplanfeatureResults, err = searchEntitlementPlanFeatures(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entityResults, err = searchEntities(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			entitytypeResults, err = searchEntityTypes(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			eventResults, err = searchEvents(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			featureResults, err = searchFeatures(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			fileResults, err = searchFiles(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			groupResults, err = searchGroups(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			groupsettingResults, err = searchGroupSettings(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			integrationResults, err = searchIntegrations(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			oauthproviderResults, err = searchOauthProviders(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			ohauthtootokenResults, err = searchOhAuthTooTokens(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			organizationResults, err = searchOrganizations(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			organizationsettingResults, err = searchOrganizationSettings(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			personalaccesstokenResults, err = searchPersonalAccessTokens(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			subscriberResults, err = searchSubscribers(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			tfasettingResults, err = searchTFASettings(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			templateResults, err = searchTemplates(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			userResults, err = searchUsers(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			usersettingResults, err = searchUserSettings(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
		func() {
			var err error
			webhookResults, err = searchWebhooks(ctx, query)
			if err != nil {
				errors = append(errors, err)
			}
		},
	})

	// Check all errors and return a single error if any of the searches failed
	if len(errors) > 0 {
		log.Error().Errs("errors", errors).Msg("search failed")

		return nil, ErrSearchFailed
	}

	// return the results
	return &SearchResultConnection{
		Nodes: []SearchResult{
			APITokenSearchResult{
				APITokens: apitokenResults,
			},
			ContactSearchResult{
				Contacts: contactResults,
			},
			DocumentDataSearchResult{
				DocumentData: documentdataResults,
			},
			EntitlementSearchResult{
				Entitlements: entitlementResults,
			},
			EntitlementPlanSearchResult{
				EntitlementPlans: entitlementplanResults,
			},
			EntitlementPlanFeatureSearchResult{
				EntitlementPlanFeatures: entitlementplanfeatureResults,
			},
			EntitySearchResult{
				Entities: entityResults,
			},
			EntityTypeSearchResult{
				EntityTypes: entitytypeResults,
			},
			EventSearchResult{
				Events: eventResults,
			},
			FeatureSearchResult{
				Features: featureResults,
			},
			FileSearchResult{
				Files: fileResults,
			},
			GroupSearchResult{
				Groups: groupResults,
			},
			GroupSettingSearchResult{
				GroupSettings: groupsettingResults,
			},
			IntegrationSearchResult{
				Integrations: integrationResults,
			},
			OauthProviderSearchResult{
				OauthProviders: oauthproviderResults,
			},
			OhAuthTooTokenSearchResult{
				OhAuthTooTokens: ohauthtootokenResults,
			},
			OrganizationSearchResult{
				Organizations: organizationResults,
			},
			OrganizationSettingSearchResult{
				OrganizationSettings: organizationsettingResults,
			},
			PersonalAccessTokenSearchResult{
				PersonalAccessTokens: personalaccesstokenResults,
			},
			SubscriberSearchResult{
				Subscribers: subscriberResults,
			},
			TFASettingSearchResult{
				TFASettings: tfasettingResults,
			},
			TemplateSearchResult{
				Templates: templateResults,
			},
			UserSearchResult{
				Users: userResults,
			},
			UserSettingSearchResult{
				UserSettings: usersettingResults,
			},
			WebhookSearchResult{
				Webhooks: webhookResults,
			},
		},
	}, nil
}
func (r *queryResolver) AdminAPITokenSearch(ctx context.Context, query string) (*APITokenSearchResult, error) {
	apitokenResults, err := adminSearchAPITokens(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &APITokenSearchResult{
		APITokens: apitokenResults,
	}, nil
}
func (r *queryResolver) AdminContactSearch(ctx context.Context, query string) (*ContactSearchResult, error) {
	contactResults, err := adminSearchContacts(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &ContactSearchResult{
		Contacts: contactResults,
	}, nil
}
func (r *queryResolver) AdminDocumentDataSearch(ctx context.Context, query string) (*DocumentDataSearchResult, error) {
	documentdataResults, err := adminSearchDocumentData(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &DocumentDataSearchResult{
		DocumentData: documentdataResults,
	}, nil
}
func (r *queryResolver) AdminEntitlementSearch(ctx context.Context, query string) (*EntitlementSearchResult, error) {
	entitlementResults, err := adminSearchEntitlements(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntitlementSearchResult{
		Entitlements: entitlementResults,
	}, nil
}
func (r *queryResolver) AdminEntitlementPlanSearch(ctx context.Context, query string) (*EntitlementPlanSearchResult, error) {
	entitlementplanResults, err := adminSearchEntitlementPlans(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntitlementPlanSearchResult{
		EntitlementPlans: entitlementplanResults,
	}, nil
}
func (r *queryResolver) AdminEntitlementPlanFeatureSearch(ctx context.Context, query string) (*EntitlementPlanFeatureSearchResult, error) {
	entitlementplanfeatureResults, err := adminSearchEntitlementPlanFeatures(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntitlementPlanFeatureSearchResult{
		EntitlementPlanFeatures: entitlementplanfeatureResults,
	}, nil
}
func (r *queryResolver) AdminEntitySearch(ctx context.Context, query string) (*EntitySearchResult, error) {
	entityResults, err := adminSearchEntities(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntitySearchResult{
		Entities: entityResults,
	}, nil
}
func (r *queryResolver) AdminEntityTypeSearch(ctx context.Context, query string) (*EntityTypeSearchResult, error) {
	entitytypeResults, err := adminSearchEntityTypes(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EntityTypeSearchResult{
		EntityTypes: entitytypeResults,
	}, nil
}
func (r *queryResolver) AdminEventSearch(ctx context.Context, query string) (*EventSearchResult, error) {
	eventResults, err := adminSearchEvents(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &EventSearchResult{
		Events: eventResults,
	}, nil
}
func (r *queryResolver) AdminFeatureSearch(ctx context.Context, query string) (*FeatureSearchResult, error) {
	featureResults, err := adminSearchFeatures(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &FeatureSearchResult{
		Features: featureResults,
	}, nil
}
func (r *queryResolver) AdminFileSearch(ctx context.Context, query string) (*FileSearchResult, error) {
	fileResults, err := adminSearchFiles(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &FileSearchResult{
		Files: fileResults,
	}, nil
}
func (r *queryResolver) AdminGroupSearch(ctx context.Context, query string) (*GroupSearchResult, error) {
	groupResults, err := adminSearchGroups(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &GroupSearchResult{
		Groups: groupResults,
	}, nil
}
func (r *queryResolver) AdminGroupSettingSearch(ctx context.Context, query string) (*GroupSettingSearchResult, error) {
	groupsettingResults, err := adminSearchGroupSettings(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &GroupSettingSearchResult{
		GroupSettings: groupsettingResults,
	}, nil
}
func (r *queryResolver) AdminIntegrationSearch(ctx context.Context, query string) (*IntegrationSearchResult, error) {
	integrationResults, err := adminSearchIntegrations(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &IntegrationSearchResult{
		Integrations: integrationResults,
	}, nil
}
func (r *queryResolver) AdminOauthProviderSearch(ctx context.Context, query string) (*OauthProviderSearchResult, error) {
	oauthproviderResults, err := adminSearchOauthProviders(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &OauthProviderSearchResult{
		OauthProviders: oauthproviderResults,
	}, nil
}
func (r *queryResolver) AdminOhAuthTooTokenSearch(ctx context.Context, query string) (*OhAuthTooTokenSearchResult, error) {
	ohauthtootokenResults, err := adminSearchOhAuthTooTokens(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &OhAuthTooTokenSearchResult{
		OhAuthTooTokens: ohauthtootokenResults,
	}, nil
}
func (r *queryResolver) AdminOrganizationSearch(ctx context.Context, query string) (*OrganizationSearchResult, error) {
	organizationResults, err := adminSearchOrganizations(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &OrganizationSearchResult{
		Organizations: organizationResults,
	}, nil
}
func (r *queryResolver) AdminOrganizationSettingSearch(ctx context.Context, query string) (*OrganizationSettingSearchResult, error) {
	organizationsettingResults, err := adminSearchOrganizationSettings(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &OrganizationSettingSearchResult{
		OrganizationSettings: organizationsettingResults,
	}, nil
}
func (r *queryResolver) AdminPersonalAccessTokenSearch(ctx context.Context, query string) (*PersonalAccessTokenSearchResult, error) {
	personalaccesstokenResults, err := adminSearchPersonalAccessTokens(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &PersonalAccessTokenSearchResult{
		PersonalAccessTokens: personalaccesstokenResults,
	}, nil
}
func (r *queryResolver) AdminSubscriberSearch(ctx context.Context, query string) (*SubscriberSearchResult, error) {
	subscriberResults, err := adminSearchSubscribers(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &SubscriberSearchResult{
		Subscribers: subscriberResults,
	}, nil
}
func (r *queryResolver) AdminTFASettingSearch(ctx context.Context, query string) (*TFASettingSearchResult, error) {
	tfasettingResults, err := adminSearchTFASettings(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &TFASettingSearchResult{
		TFASettings: tfasettingResults,
	}, nil
}
func (r *queryResolver) AdminTemplateSearch(ctx context.Context, query string) (*TemplateSearchResult, error) {
	templateResults, err := adminSearchTemplates(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &TemplateSearchResult{
		Templates: templateResults,
	}, nil
}
func (r *queryResolver) AdminUserSearch(ctx context.Context, query string) (*UserSearchResult, error) {
	userResults, err := adminSearchUsers(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &UserSearchResult{
		Users: userResults,
	}, nil
}
func (r *queryResolver) AdminUserSettingSearch(ctx context.Context, query string) (*UserSettingSearchResult, error) {
	usersettingResults, err := adminSearchUserSettings(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &UserSettingSearchResult{
		UserSettings: usersettingResults,
	}, nil
}
func (r *queryResolver) AdminWebhookSearch(ctx context.Context, query string) (*WebhookSearchResult, error) {
	webhookResults, err := adminSearchWebhooks(ctx, query)

	if err != nil {
		return nil, ErrSearchFailed
	}

	// return the results
	return &WebhookSearchResult{
		Webhooks: webhookResults,
	}, nil
}
