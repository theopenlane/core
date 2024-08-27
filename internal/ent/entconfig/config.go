package entconfig

// Config holds the configuration for the ent server
type Config struct {
	// Flags contains the flags for the server to allow use to test different code paths
	Flags Flags `json:"flags" koanf:"flags" jsonschema:"description=flags for the server"`
	// EntityTypes is the list of entity types to create by default for the organization
	EntityTypes []string `json:"entityTypes" koanf:"entityTypes" default:"" description:"entity types to create for the organization"`
}

// Flags contains the flags for the server to allow use to test different code paths
type Flags struct {
	// UseListUserService is a flag to use the list user service endpoint for users with object access, if false, the core db is used directly instead
	UseListUserService bool `json:"useListUserService" koanf:"useListUserService" jsonschema:"description=use list services endpoint for object access" default:"true"`
	// UserListObjectService is a flag to use the list object service endpoint for object access for a user, if false, the core db is used directly instead
	UseListObjectService bool `json:"useListObjectServices" koanf:"useListObjectServices" jsonschema:"description=use list object services endpoint for object access" default:"false"`
}
