package catalog

import (
	"log"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
	c "github.com/theopenlane/core/pkg/catalog"
)

func main() {
	tmpl := &c.Catalog{
		Modules: map[string]c.Feature{
			"compliance": {
				DisplayName: "Compliance",
				Description: "comply it up big dawg",
				Audience:    "public",
				Billing: c.Billing{
					Prices: []c.Price{
						{
							Interval:   "month",
							UnitAmount: 1000,
						},
					},
				},
			},
		},
		Addons: map[string]c.Feature{
			"vanity domain": {
				DisplayName: "Vanity Domain",
				Description: "Serve your UI on custom hostname.",
				Audience:    "beta",
				Billing: c.Billing{
					Prices: []c.Price{
						{
							Interval:   "month",
							UnitAmount: 500,
						},
					},
				},
			},
		},
	}

	out, err := yaml.Marshal(tmpl)
	if err != nil {
		log.Fatal(err)
	}

	path := filepath.Join("..", "..", "..", "catalog-config", "config", "catalog.scaffold.yaml")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		log.Fatal(err)
	}
}
