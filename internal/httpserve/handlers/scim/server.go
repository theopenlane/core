package scim

import (
	"github.com/elimity-com/scim"
	opt "github.com/elimity-com/scim/optional"
	"github.com/elimity-com/scim/schema"

	integrationsruntime "github.com/theopenlane/core/internal/integrations/runtime"
)

const (
	// maxSCIMResults is the maximum number of results returned in a SCIM list response
	maxSCIMResults = 200
)

// NewSCIMServer creates a new SCIM server with User and Group resource handlers
func NewSCIMServer(rt *integrationsruntime.Runtime) (scim.Server, error) {
	config := scim.ServiceProviderConfig{
		DocumentationURI: opt.NewString("https://docs.theopenlane.io/scim"),
		AuthenticationSchemes: []scim.AuthenticationScheme{
			{
				Type:        scim.AuthenticationTypeOauthBearerToken,
				Name:        "SCIM Bearer Token",
				Description: "Authenticate with the SCIM bearer secret provisioned for this installation",
			},
		},
		SupportPatch:     true,
		SupportFiltering: false,
		MaxResults:       maxSCIMResults,
	}

	userResourceType := scim.ResourceType{
		ID:          opt.NewString("User"),
		Name:        "User",
		Description: opt.NewString("Directory user account"),
		Endpoint:    "/Users",
		Schema:      directoryUserSchema(),
		Handler:     &DirectoryUserHandler{Runtime: rt},
	}

	groupResourceType := scim.ResourceType{
		ID:          opt.NewString("Group"),
		Name:        "Group",
		Description: opt.NewString("Directory group resource"),
		Endpoint:    "/Groups",
		Schema:      directoryGroupSchema(),
		Handler:     &DirectoryGroupHandler{Runtime: rt},
	}

	return scim.NewServer(
		&scim.ServerArgs{
			ServiceProviderConfig: &config,
			ResourceTypes: []scim.ResourceType{
				userResourceType,
				groupResourceType,
			},
		},
	)
}

// directoryUserSchema returns a composed SCIM User schema containing only
// the attributes our directory account handlers actually process
func directoryUserSchema() schema.Schema {
	return schema.Schema{
		ID:          schema.UserSchema,
		Name:        opt.NewString("User"),
		Description: opt.NewString("Directory user account"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: opt.NewString("Unique identifier for the User, typically the email address"),
				Name:        "userName",
				Required:    true,
				Uniqueness:  schema.AttributeUniquenessServer(),
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: opt.NewString("The components of the user's real name"),
				Name:        "name",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: opt.NewString("The given name of the User"),
						Name:        "givenName",
					}),
					schema.SimpleStringParams(schema.StringParams{
						Description: opt.NewString("The family name of the User"),
						Name:        "familyName",
					}),
				},
			}),
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: opt.NewString("The name of the User, suitable for display"),
				Name:        "displayName",
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Description: opt.NewString("A Boolean value indicating the User's administrative status"),
				Name:        "active",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: opt.NewString("Email addresses for the user"),
				MultiValued: true,
				Name:        "emails",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: opt.NewString("Email address value"),
						Name:        "value",
					}),
					schema.SimpleBooleanParams(schema.BooleanParams{
						Description: opt.NewString("Indicates the primary email address"),
						Name:        "primary",
					}),
				},
			}),
		},
	}
}

// directoryGroupSchema returns a composed SCIM Group schema containing only
// the attributes our directory group handlers actually process
func directoryGroupSchema() schema.Schema {
	return schema.Schema{
		ID:          schema.GroupSchema,
		Name:        opt.NewString("Group"),
		Description: opt.NewString("Directory group resource"),
		Attributes: []schema.CoreAttribute{
			schema.SimpleCoreAttribute(schema.SimpleStringParams(schema.StringParams{
				Description: opt.NewString("A human-readable name for the Group"),
				Name:        "displayName",
				Required:    true,
			})),
			schema.SimpleCoreAttribute(schema.SimpleBooleanParams(schema.BooleanParams{
				Description: opt.NewString("A Boolean value indicating the Group's active status"),
				Name:        "active",
			})),
			schema.ComplexCoreAttribute(schema.ComplexParams{
				Description: opt.NewString("A list of members of the Group"),
				MultiValued: true,
				Name:        "members",
				SubAttributes: []schema.SimpleParams{
					schema.SimpleStringParams(schema.StringParams{
						Description: opt.NewString("Identifier of the member of this Group"),
						Mutability:  schema.AttributeMutabilityImmutable(),
						Name:        "value",
					}),
					schema.SimpleReferenceParams(schema.ReferenceParams{
						Description:    opt.NewString("The URI of the member resource"),
						Mutability:     schema.AttributeMutabilityImmutable(),
						Name:           "$ref",
						ReferenceTypes: []schema.AttributeReferenceType{"User"},
					}),
				},
			}),
		},
	}
}
