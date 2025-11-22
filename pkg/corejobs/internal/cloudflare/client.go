package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v6"
	"github.com/cloudflare/cloudflare-go/v6/custom_hostnames"
	"github.com/cloudflare/cloudflare-go/v6/dns"
	"github.com/cloudflare/cloudflare-go/v6/option"
	"github.com/cloudflare/cloudflare-go/v6/packages/pagination"
	"github.com/cloudflare/cloudflare-go/v6/zones"
)

// CustomHostnamesService defines the interface for interacting with Cloudflare's
// custom hostnames API. It provides methods for creating and deleting custom hostnames.
type CustomHostnamesService interface {
	// New creates a new custom hostname in Cloudflare with the provided parameters.
	New(context.Context, custom_hostnames.CustomHostnameNewParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameNewResponse, error)
	// Delete removes a custom hostname from Cloudflare using the provided hostname ID and parameters.
	Delete(context.Context, string, custom_hostnames.CustomHostnameDeleteParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameDeleteResponse, error)
	// Get retrieves a custom hostname from Cloudflare using the provided hostname ID and parameters.
	Get(context.Context, string, custom_hostnames.CustomHostnameGetParams, ...option.RequestOption) (*custom_hostnames.CustomHostnameGetResponse, error)
}

type ZonesService interface {
	Get(ctx context.Context, query zones.ZoneGetParams, opts ...option.RequestOption) (res *zones.Zone, err error)
}

type RecordService interface {
	New(ctx context.Context, params dns.RecordNewParams, opts ...option.RequestOption) (res *dns.RecordResponse, err error)
	List(ctx context.Context, params dns.RecordListParams, opts ...option.RequestOption) (res *pagination.V4PagePaginationArray[dns.RecordResponse], err error)
}

// Client defines the interface for the Cloudflare client.
// It provides access to various Cloudflare API services.
type Client interface {
	// CustomHostnames returns the service for managing custom hostnames in Cloudflare.
	CustomHostnames() CustomHostnamesService
	// Zones returns the service for managing zones in Cloudflare.
	Zones() ZonesService
	// DNS returns the service for managing DNS records in Cloudflare.
	Record() RecordService
}

// cloudflareClientImpl implements the Client interface using the official Cloudflare Go client.
type cloudflareClientImpl struct {
	client *cloudflare.Client
}

// NewClient creates a new Cloudflare client using the provided API key.
// It returns an implementation of the Client interface.
func NewClient(apiKey string) Client {
	return &cloudflareClientImpl{
		client: cloudflare.NewClient(
			option.WithAPIToken(apiKey),
		),
	}
}

// CustomHostnames returns the service for managing custom hostnames in Cloudflare.
// It implements the Client interface method.
func (c *cloudflareClientImpl) CustomHostnames() CustomHostnamesService {
	return c.client.CustomHostnames
}

func (c *cloudflareClientImpl) Zones() ZonesService {
	return c.client.Zones
}

func (c *cloudflareClientImpl) Record() RecordService {
	return c.client.DNS.Records
}
