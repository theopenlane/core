package plan

import (
	"time"

	"github.com/theopenlane/core/pkg/billing/product"
)

// Plan is a collection of products it is a logical grouping of products and doesn't have a corresponding billing engine entity
type Plan struct {
	ID          string `json:"id" yaml:"id"`
	Name        string `json:"name" yaml:"name"`
	Title       string `json:"title" yaml:"title"`
	Description string `json:"description" yaml:"description"`
	// Interval is the interval at which the plan is billed
	Interval string `json:"interval" yaml:"interval"`
	// OnStartCredits is the number of credits that are awarded when a subscription is started
	OnStartCredits int64 `json:"on_start_credits" yaml:"on_start_credits"`
	// Products for the plan, return only, should not be set when creating a plan
	Products []product.Product `json:"products" yaml:"products"`
	// TrialDays is the number of days a subscription is in trial
	TrialDays int64 `json:"trial_days" yaml:"trial_days"`
	State     string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Filter is a filter for plans
type Filter struct {
	IDs      []string
	Interval string
	State    string
}

// File is a grouping struct
type File struct {
	Plans    []Plan            `json:"plans" yaml:"plans"`
	Products []product.Product `json:"products" yaml:"products"`
	Features []product.Feature `json:"features" yaml:"features"`
}
