package catalog

import (
	"context"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/theopenlane/core/pkg/entitlements"
)

// Billing describes pricing details for a module or addon.
type Billing struct {
	Interval   string `json:"interval" yaml:"interval"`
	UnitAmount int64  `json:"unit_amount" yaml:"unit_amount"`
	PriceID    string `json:"price_id" yaml:"price_id"`
}

// Feature defines a purchasable module or addon feature.
type Feature struct {
	DisplayName string  `json:"display_name" yaml:"display_name"`
	Description string  `json:"description,omitempty" yaml:"description,omitempty"`
	Billing     Billing `json:"billing" yaml:"billing"`
	Audience    string  `json:"audience" yaml:"audience"`
}

// FeatureSet is a mapping of feature identifiers to metadata.
type FeatureSet map[string]Feature

// Catalog contains all modules and addons offered by Openlane.
type Catalog struct {
	Modules FeatureSet `json:"modules" yaml:"modules"`
	Addons  FeatureSet `json:"addons" yaml:"addons"`
}

// Visible returns modules and addons filtered by audience.
// Providing "" returns everything.
func (c *Catalog) Visible(audience string) *Catalog {
	if c == nil {
		return &Catalog{}
	}

	filter := func(in FeatureSet) FeatureSet {
		if audience == "" {
			return in
		}
		out := FeatureSet{}
		for k, v := range in {
			if v.Audience == audience || v.Audience == "public" {
				out[k] = v
			}
		}
		return out
	}

	return &Catalog{
		Modules: filter(c.Modules),
		Addons:  filter(c.Addons),
	}
}

// LoadCatalog reads and parses a Catalog definition from disk.
func LoadCatalog(path string) (*Catalog, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		return nil, fs.ErrNotExist
	}

	var c Catalog
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &c, nil
}

// ValidatePrices ensures every feature with a price_id exists in Stripe.
func (c *Catalog) ValidatePrices(ctx context.Context, sc *entitlements.StripeClient) error {
	if c == nil || sc == nil {
		return nil
	}

	check := func(fs FeatureSet) error {
		for name, f := range fs {
			if f.Billing.PriceID == "" {
				continue
			}
			if _, err := sc.GetPrice(ctx, f.Billing.PriceID); err != nil {
				return fmt.Errorf("price %s for feature %s not found: %w", f.Billing.PriceID, name, err)
			}
		}
		return nil
	}

	if err := check(c.Modules); err != nil {
		return err
	}
	return check(c.Addons)
}

// EnsurePrices verifies prices exist in Stripe and creates them when missing.
// New products are created using the feature display name and description.
// The created price ID is stored back in the catalog struct but not persisted
// to disk.
func (c *Catalog) EnsurePrices(ctx context.Context, sc *entitlements.StripeClient, currency string) error {
	if c == nil || sc == nil {
		return nil
	}

	create := func(name string, f Feature) (Feature, error) {
		prod, err := sc.CreateProduct(ctx, f.DisplayName, f.Description)
		if err != nil {
			return f, fmt.Errorf("create product for %s: %w", name, err)
		}
		price, err := sc.CreatePrice(ctx, prod.ID, f.Billing.UnitAmount, currency, f.Billing.Interval)
		if err != nil {
			return f, fmt.Errorf("create price for %s: %w", name, err)
		}
		f.Billing.PriceID = price.ID
		return f, nil
	}

	ensure := func(fs FeatureSet) error {
		for name, feat := range fs {
			if feat.Billing.PriceID != "" {
				if _, err := sc.GetPrice(ctx, feat.Billing.PriceID); err == nil {
					continue
				}
			}

			var err error
			feat, err = create(name, feat)
			if err != nil {
				return err
			}
			fs[name] = feat
		}
		return nil
	}

	if err := ensure(c.Modules); err != nil {
		return err
	}
	return ensure(c.Addons)
}
