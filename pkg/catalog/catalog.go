package catalog

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"fmt"
	"io/fs"
	"maps"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/rs/zerolog/log"

	"github.com/stripe/stripe-go/v84"
	"github.com/xeipuuv/gojsonschema"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/entitlements"
)

//go:embed genjsonschema/catalog.schema.json
var schemaBytes []byte

// ManagedByKey is the metadata key applied to Stripe resources created via the catalog
const ManagedByKey = "managed_by"

// ManagedByValue identifies objects managed by the catalog automation
const ManagedByValue = "module-manager"

// Catalog wraps models.Catalog
type Catalog struct {
	// embed the base catalog model
	models.Catalog
}

type (
	FeatureSet = models.FeatureSet
	Feature    = models.Feature
	Billing    = models.Billing
	Price      = models.ItemPrice
	Usage      = models.Usage
)

// New creates a new empty Catalog instance
func New() *Catalog {
	return &Catalog{}
}

// computeSHA returns the hex encoded SHA256 of the provided version string.
func computeSHA(ver string) string {
	sum := sha256.Sum256([]byte(ver))
	return hex.EncodeToString(sum[:])
}

// IsCurrent reports whether the catalog SHA matches its version.
func (c *Catalog) IsCurrent() bool {
	if c == nil {
		return true
	}

	return c.SHA == computeSHA(c.Version)
}

// Visible returns modules and addons filtered by audience
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

	visible := New()

	visible.Modules = filter(c.Modules)
	visible.Addons = filter(c.Addons)

	return visible
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

	schema := gojsonschema.NewBytesLoader(schemaBytes)

	jsonData, err := yaml.YAMLToJSON(data)
	if err != nil {
		return nil, ErrYamlToJSONConversion
	}

	doc := gojsonschema.NewBytesLoader(jsonData)

	// this effectively "lints" the catalog to ensure it conforms against the schema
	res, err := gojsonschema.Validate(schema, doc)
	if err != nil {
		return nil, err
	}

	if !res.Valid() {
		log.Debug().Msg("Catalog validation failed - ensure you have generated the latest schema file if you have modified the catalog structs")

		// if there are errors with many fields its easiest to see them this way
		for _, e := range res.Errors() {
			log.Debug().Msgf("Validation error: %s", e)
		}

		return nil, ErrCatalogValidationFailed
	}

	var c models.Catalog

	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	return &Catalog{Catalog: c}, nil
}

// ValidatePrices ensures every feature's price attributes match a Stripe price.
// Matching considers unit amount, interval, nickname, lookup key and metadata.
func (c *Catalog) ValidatePrices(ctx context.Context, sc *entitlements.StripeClient) error {
	prodMap, err := c.getProductMap(ctx, sc)
	if err != nil {
		return err
	}

	check := func(fs FeatureSet) error {
		for name, f := range fs {
			prod, _ := resolveProduct(ctx, sc, prodMap, f)
			if prod == nil {
				return ErrProductMissingFeature
			}

			for i, p := range f.Billing.Prices {
				md := map[string]string{ManagedByKey: ManagedByValue}

				maps.Copy(md, p.Metadata)

				price, err := sc.FindPriceForProduct(ctx, prod.ID, p.PriceID, p.UnitAmount, "", p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					return err
				}

				if price == nil {
					return ErrMatchingPriceNotFound
				}

				f.Billing.Prices[i].PriceID = price.ID
			}

			f.ProductID = prod.ID

			fs[name] = f
		}

		return nil
	}

	if err := check(c.Modules); err != nil {
		return err
	}

	return check(c.Addons)
}

func (c *Catalog) getProductMap(ctx context.Context, sc *entitlements.StripeClient) (map[string]*stripe.Product, error) {
	if c == nil || sc == nil {
		return nil, ErrContextandClientRequired
	}

	products, err := sc.ListProducts(ctx)
	if err != nil {
		log.Error().Err(err).Msg("failed to list products from Stripe")
		return nil, err
	}

	prodMap := map[string]*stripe.Product{}

	for _, p := range products {
		if p.ID != "" {
			prodMap[p.ID] = p
		}

		if p.Name != "" {
			prodMap[p.Name] = p
		}
	}

	return prodMap, nil
}

// EnsurePrices verifies prices exist in Stripe and creates them when missing.
// New products are created using the feature display name and description.
// Matching is performed by unit amount, interval, nickname, lookup key and
// metadata instead of a fixed price ID. The discovered Stripe price ID is stored back in the catalog
// struct but not persisted to disk.
func (c *Catalog) EnsurePrices(ctx context.Context, sc *entitlements.StripeClient, currency string) error {
	prodMap, err := c.getProductMap(ctx, sc)
	if err != nil {
		return err
	}

	create := func(name string, f Feature) (Feature, error) {
		log.Info().Str("feature", name).Msg("ensuring feature prices exist in Stripe")

		prod, _ := resolveProduct(ctx, sc, prodMap, f)

		lookup := name
		if f.LookupKey != "" {
			lookup = f.LookupKey
		}

		var feature *stripe.EntitlementsFeature

		feature, err = sc.GetFeatureByLookupKey(ctx, lookup)
		if err != nil {
			log.Err(err).Msg("failed to get feature by lookup key")
			return f, err
		}

		if prod == nil {
			metadata := map[string]string{
				ManagedByKey: ManagedByValue,
				"module":     name,
			}

			prod, err = sc.CreateProduct(ctx, f.DisplayName, f.Description, metadata)
			if err != nil {
				log.Error().Err(err).Msg("failed to create product")
				return f, ErrFailedToCreateProduct
			}

			prodMap[f.DisplayName] = prod
		} else if prod.Metadata == nil || prod.Metadata["module"] == "" {
			// Product exists, check if it has the module metadata
			// Need to update the product metadata
			existingMetadata := make(map[string]string)

			if prod.Metadata != nil {
				for k, v := range prod.Metadata {
					existingMetadata[k] = v
				}
			}

			// Add missing metadata
			existingMetadata[ManagedByKey] = ManagedByValue
			existingMetadata["module"] = name

			// Update the product with the new metadata
			updateParams := &stripe.ProductUpdateParams{}
			updateParams = sc.UpdateProductWithOptions(updateParams,
				entitlements.WithUpdateProductMetadata(existingMetadata))

			updatedProd, err := sc.UpdateProductWithParams(ctx, prod.ID, updateParams)
			if err != nil {
				log.Error().Err(err).Msg("failed to update product metadata")
				return f, fmt.Errorf("failed to update product metadata: %w", err)
			}

			prod = updatedProd
			prodMap[f.DisplayName] = prod
		}

		if feature == nil {
			log.Info().Str("feature", name).Msg("creating entitlement feature in Stripe")

			feature, err = sc.CreateProductFeatureWithOptions(ctx,
				&stripe.EntitlementsFeatureCreateParams{},
				entitlements.WithFeatureName(f.DisplayName),
				entitlements.WithFeatureLookupKey(lookup),
			)
			if err != nil {
				log.Error().Err(err).Msg("failed to create entitlement feature")
				return f, err
			}

			log.Info().Str("feature", name).Str("id", feature.ID).Msg("created entitlement feature in Stripe")
			_, _ = sc.AttachFeatureToProductWithOptions(ctx,
				&stripe.ProductFeatureCreateParams{},
				entitlements.WithProductFeatureProductID(prod.ID),
				entitlements.WithProductFeatureEntitlementFeatureID(feature.ID),
			)
		}

		f.ProductID = prod.ID

		var monthPriceID, yearPriceID string

		for i, p := range f.Billing.Prices {
			md := map[string]string{ManagedByKey: ManagedByValue}

			maps.Copy(md, p.Metadata)

			price, err := sc.FindPriceForProduct(ctx, prod.ID, p.PriceID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
			if err != nil && !notFound(err) {
				return f, err
			}

			if price == nil {
				log.Info().Str("feature", name).Msg("creating missing price in Stripe")

				price, err = sc.CreatePrice(ctx, prod.ID, p.UnitAmount, currency, p.Interval, p.Nickname, p.LookupKey, md)
				if err != nil {
					log.Error().Err(err).Msg("failed to create price")
					return f, ErrFailedToCreatePrice
				}
			}

			f.Billing.Prices[i].PriceID = price.ID

			switch p.Interval {
			case "month":
				monthPriceID = price.ID
			case "year":
				yearPriceID = price.ID
			}
		}

		if monthPriceID != "" && yearPriceID != "" {
			if err := sc.TagPriceUpsell(ctx, monthPriceID, yearPriceID); err != nil {
				log.Error().Err(err).Msg("failed to tag price upsell")
				return f, err
			}
		}

		return f, nil
	}

	ensure := func(fs FeatureSet) error {
		for name, feat := range fs {
			var err error

			feat, err = create(name, feat)
			if err != nil {
				log.Error().Err(err).Msg("failed to create feature")
				return err
			}

			fs[name] = feat
		}

		return nil
	}

	if err := ensure(c.Modules); err != nil {
		log.Error().Err(err).Msg("failed to ensure module features")
		return err
	}

	return ensure(c.Addons)
}

// SaveCatalog writes the catalog to disk in YAML format, as well as computing and updating the SHA
func (c *Catalog) SaveCatalog(path string) (string, error) {
	if c == nil {
		return "", nil
	}

	origData, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	var orig Catalog
	if len(origData) > 0 {
		if err := yaml.Unmarshal(origData, &orig); err != nil {
			return "", err
		}
	}

	// carry over version if not set and ensure SHA present
	if c.Version == "" && orig.Version != "" {
		c.Version = orig.Version
	}

	if c.SHA == "" {
		c.SHA = computeSHA(c.Version)
	}

	newData, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}

	diff := ""

	if !bytes.Equal(origData, newData) {
		ver := c.Version
		if ver == "" {
			ver = models.DefaultRevision
		}

		if bumped, berr := models.BumpPatch(ver); berr == nil {
			c.Version = bumped

			c.SHA = computeSHA(c.Version)

			if newData, err = yaml.Marshal(c); err != nil {
				return "", err
			}
		}

		u := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(origData)),
			B:        difflib.SplitLines(string(newData)),
			FromFile: "catalog(old)",
			ToFile:   "catalog(new)",
			Context:  3, //nolint:mnd
		}

		diff, _ = difflib.GetUnifiedDiffString(u)
	}

	if err := os.WriteFile(path, newData, 0o644); err != nil { //nolint:gosec,mnd
		return "", err
	}

	return diff, nil
}

// resolveProduct attempts to locate the Stripe product for a feature using the
// most specific information available. It tries any referenced price IDs first,
// then lookup keys, and finally falls back to the feature display name.
func resolveProduct(ctx context.Context, sc *entitlements.StripeClient, prodMap map[string]*stripe.Product, feat Feature) (*stripe.Product, error) {
	for _, p := range feat.Billing.Prices {
		if p.PriceID != "" {
			pr, err := sc.GetPrice(ctx, p.PriceID)
			if err == nil && pr != nil && pr.Product != nil {
				return sc.GetProductByID(ctx, pr.Product.ID)
			}
		}

		if p.LookupKey != "" {
			pr, err := sc.GetPriceByLookupKey(ctx, p.LookupKey)
			if err == nil && pr != nil && pr.Product != nil {
				return sc.GetProductByID(ctx, pr.Product.ID)
			}
		}
	}

	if prod, ok := prodMap[feat.DisplayName]; ok {
		return prod, nil
	}

	return nil, nil
}
