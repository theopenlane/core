package catalog

import "fmt"

var validIntervals = map[string]struct{}{
	"day":   {},
	"week":  {},
	"month": {},
	"year":  {},
}

// Lint validates catalog contents for basic correctness.
func (c *Catalog) Lint() error {
	if c == nil {
		return nil
	}
	check := func(fs FeatureSet) error {
		for name, f := range fs {
			for _, p := range f.Billing.Prices {
				if _, ok := validIntervals[p.Interval]; !ok {
					return fmt.Errorf("invalid interval %s for feature %s", p.Interval, name)
				}
			}
		}
		return nil
	}
	if err := check(c.Modules); err != nil {
		return err
	}
	return check(c.Addons)
}
