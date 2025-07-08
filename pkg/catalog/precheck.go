package catalog

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v82"
)

// LookupKeyConflict describes an existing Stripe price using a lookup key.
type LookupKeyConflict struct {
	Feature   string
	Resource  string
	LookupKey string
	ID        string
}

// lookupKeyCheckConfig holds optional settings for checking lookup keys.
type lookupKeyCheckConfig struct {
	failFast bool
}

// LookupKeyCheckOption configures lookup key conflict checks.
type LookupKeyCheckOption func(*lookupKeyCheckConfig)

// WithFailFast stops checking after the first conflict is found.
func WithFailFast(f bool) LookupKeyCheckOption {
	return func(c *lookupKeyCheckConfig) { c.failFast = f }
}

// LookupKeyConflicts scans all feature prices and reports lookup keys that
// already exist in Stripe under a different price ID.
type lookupPriceGetter interface {
	GetPriceByLookupKey(ctx context.Context, lookupKey string) (*stripe.Price, error)
}

type lookupFeatureGetter interface {
	GetFeatureByLookupKey(ctx context.Context, lookupKey string) (*stripe.EntitlementsFeature, error)
}

type lookupProductGetter interface {
	GetProduct(ctx context.Context, productID string) (*stripe.Product, error)
}

type lookupClient interface {
	lookupPriceGetter
	lookupFeatureGetter
	lookupProductGetter
}

func (c *Catalog) LookupKeyConflicts(ctx context.Context, sc lookupClient, opts ...LookupKeyCheckOption) ([]LookupKeyConflict, error) {
	if c == nil || sc == nil {
		return nil, ErrContextandClientRequired
	}

	cfg := &lookupKeyCheckConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	var conflicts []LookupKeyConflict

	check := func(kind string, fs FeatureSet) error {
		for name, f := range fs {
			if prod, err := sc.GetProduct(ctx, name); err != nil {
				return err
			} else if prod != nil {
				conflicts = append(conflicts, LookupKeyConflict{Feature: fmt.Sprintf("%s:%s", kind, name), Resource: "product", LookupKey: name, ID: prod.ID})
				if cfg.failFast {
					return nil
				}
			}

			if f.LookupKey != "" {
				feat, err := sc.GetFeatureByLookupKey(ctx, f.LookupKey)
				if err != nil {
					return err
				}
				if feat != nil {
					conflicts = append(conflicts, LookupKeyConflict{Feature: fmt.Sprintf("%s:%s", kind, name), Resource: "feature", LookupKey: f.LookupKey, ID: feat.ID})
					if cfg.failFast {
						return nil
					}
				}
			}

			for _, p := range f.Billing.Prices {
				if p.LookupKey == "" {
					continue
				}

				price, err := sc.GetPriceByLookupKey(ctx, p.LookupKey)
				if err != nil {
					return err
				}

				if price != nil && p.PriceID != price.ID {
					conflicts = append(conflicts, LookupKeyConflict{Feature: fmt.Sprintf("%s:%s", kind, name), Resource: "price", LookupKey: p.LookupKey, ID: price.ID})
					if cfg.failFast {
						return nil
					}
				}
			}
		}
		return nil
	}

	if err := check("module", c.Modules); err != nil {
		return nil, err
	}
	if err := check("addon", c.Addons); err != nil {
		return nil, err
	}

	return conflicts, nil
}
