package reconciler

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v83"

	"github.com/theopenlane/shared/catalog"
)

func TestShouldUpdateCancelBehavior(t *testing.T) {
	tests := []struct {
		name     string
		sub      *stripe.Subscription
		expected bool
	}{
		{
			name: "subscription with pause behavior should be updated",
			sub: &stripe.Subscription{
				ID: "sub_pause",
				TrialSettings: &stripe.SubscriptionTrialSettings{
					EndBehavior: &stripe.SubscriptionTrialSettingsEndBehavior{
						MissingPaymentMethod: stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodPause,
					},
				},
			},
			expected: true,
		},
		{
			name: "subscription with cancel behavior should not be updated",
			sub: &stripe.Subscription{
				ID: "sub_cancel",
				TrialSettings: &stripe.SubscriptionTrialSettings{
					EndBehavior: &stripe.SubscriptionTrialSettingsEndBehavior{
						MissingPaymentMethod: stripe.SubscriptionTrialSettingsEndBehaviorMissingPaymentMethodCancel,
					},
				},
			},
			expected: false,
		},
		{
			name: "subscription without trial settings should not be updated",
			sub: &stripe.Subscription{
				ID:            "sub_no_trial",
				TrialSettings: nil,
			},
			expected: false,
		},
		{
			name: "subscription without end behavior should not be updated",
			sub: &stripe.Subscription{
				ID: "sub_no_end_behavior",
				TrialSettings: &stripe.SubscriptionTrialSettings{
					EndBehavior: nil,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUpdateCancelBehavior(tt.sub)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldCreateSchedule(t *testing.T) {
	tests := []struct {
		name     string
		sub      *stripe.Subscription
		expected bool
	}{
		{
			name: "subscription without schedule should have one created",
			sub: &stripe.Subscription{
				ID:       "sub_no_schedule",
				Schedule: nil,
			},
			expected: true,
		},
		{
			name: "subscription with empty schedule ID should have one created",
			sub: &stripe.Subscription{
				ID: "sub_empty_schedule",
				Schedule: &stripe.SubscriptionSchedule{
					ID: "",
				},
			},
			expected: true,
		},
		{
			name: "subscription with existing schedule should not have one created",
			sub: &stripe.Subscription{
				ID: "sub_with_schedule",
				Schedule: &stripe.SubscriptionSchedule{
					ID: "sched_existing",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldCreateSchedule(tt.sub)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateScheduleActionDescription(t *testing.T) {
	tests := []struct {
		name               string
		sub                *stripe.Subscription
		expectedAction     string
		expectedCustomerID string
	}{
		{
			name: "subscription with customer should include customer update",
			sub: &stripe.Subscription{
				ID:       "sub_123",
				Customer: &stripe.Customer{ID: "cus_456"},
			},
			expectedAction:     "create subscription schedule for sub_123 and update customer metadata for cus_456",
			expectedCustomerID: "cus_456",
		},
		{
			name: "subscription without customer should only mention schedule creation",
			sub: &stripe.Subscription{
				ID:       "sub_789",
				Customer: nil,
			},
			expectedAction:     "create subscription schedule for sub_789",
			expectedCustomerID: "",
		},
		{
			name: "subscription with empty customer ID should only mention schedule creation",
			sub: &stripe.Subscription{
				ID:       "sub_abc",
				Customer: &stripe.Customer{ID: ""},
			},
			expectedAction:     "create subscription schedule for sub_abc",
			expectedCustomerID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, customerID := GenerateScheduleActionDescription(tt.sub)
			assert.Equal(t, tt.expectedAction, action)
			assert.Equal(t, tt.expectedCustomerID, customerID)
		})
	}
}

func TestBuildValidProductsMap(t *testing.T) {
	cat := &catalog.Catalog{
		Modules: catalog.FeatureSet{
			"module1": catalog.Feature{ProductID: "prod_module_1"},
			"module2": catalog.Feature{ProductID: "prod_module_2"},
			"module3": catalog.Feature{ProductID: ""}, // Empty product ID should be ignored
		},
		Addons: catalog.FeatureSet{
			"addon1": catalog.Feature{ProductID: "prod_addon_1"},
			"addon2": catalog.Feature{ProductID: "prod_addon_2"},
			"addon3": catalog.Feature{ProductID: ""}, // Empty product ID should be ignored
		},
	}

	validProducts := BuildValidProductsMap(cat)

	expected := map[string]bool{
		"prod_module_1": true,
		"prod_module_2": true,
		"prod_addon_1":  true,
		"prod_addon_2":  true,
	}

	assert.Equal(t, expected, validProducts)
}

func TestFindMissingProductsInSubscription(t *testing.T) {
	validProducts := map[string]bool{
		"prod_valid_1": true,
		"prod_valid_2": true,
	}

	tests := []struct {
		name     string
		sub      *stripe.Subscription
		expected []SubscriptionProductReport
	}{
		{
			name: "active subscription with invalid products should be reported",
			sub: &stripe.Subscription{
				ID:       "sub_invalid",
				Status:   stripe.SubscriptionStatusActive,
				Customer: &stripe.Customer{ID: "cus_123"},
				Metadata: map[string]string{"organization_id": "org_456"},
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_1",
								Product: &stripe.Product{
									ID:   "prod_invalid_1",
									Name: "Invalid Product 1",
								},
							},
						},
						{
							Price: &stripe.Price{
								ID: "price_2",
								Product: &stripe.Product{
									ID:   "prod_invalid_2",
									Name: "Invalid Product 2",
								},
							},
						},
					},
				},
			},
			expected: []SubscriptionProductReport{
				{
					SubscriptionID: "sub_invalid",
					CustomerID:     "cus_123",
					ProductID:      "prod_invalid_1",
					ProductName:    "Invalid Product 1",
					Status:         "active",
					OrganizationID: "org_456",
				},
				{
					SubscriptionID: "sub_invalid",
					CustomerID:     "cus_123",
					ProductID:      "prod_invalid_2",
					ProductName:    "Invalid Product 2",
					Status:         "active",
					OrganizationID: "org_456",
				},
			},
		},
		{
			name: "active subscription with valid products should not be reported",
			sub: &stripe.Subscription{
				ID:     "sub_valid",
				Status: stripe.SubscriptionStatusActive,
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_3",
								Product: &stripe.Product{
									ID:   "prod_valid_1",
									Name: "Valid Product 1",
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "inactive subscription should not be reported",
			sub: &stripe.Subscription{
				ID:     "sub_canceled",
				Status: stripe.SubscriptionStatusCanceled,
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_4",
								Product: &stripe.Product{
									ID:   "prod_invalid_3",
									Name: "Invalid Product 3",
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		{
			name: "subscription with missing customer/metadata should handle gracefully",
			sub: &stripe.Subscription{
				ID:       "sub_missing_data",
				Status:   stripe.SubscriptionStatusActive,
				Customer: nil,
				Metadata: nil,
				Items: &stripe.SubscriptionItemList{
					Data: []*stripe.SubscriptionItem{
						{
							Price: &stripe.Price{
								ID: "price_5",
								Product: &stripe.Product{
									ID:   "prod_invalid_4",
									Name: "Invalid Product 4",
								},
							},
						},
					},
				},
			},
			expected: []SubscriptionProductReport{
				{
					SubscriptionID: "sub_missing_data",
					CustomerID:     "",
					ProductID:      "prod_invalid_4",
					ProductName:    "Invalid Product 4",
					Status:         "active",
					OrganizationID: "",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindMissingProductsInSubscription(tt.sub, validProducts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSubscriptionActiveOrTrialing(t *testing.T) {
	tests := []struct {
		name     string
		status   stripe.SubscriptionStatus
		expected bool
	}{
		{
			name:     "active subscription should return true",
			status:   stripe.SubscriptionStatusActive,
			expected: true,
		},
		{
			name:     "trialing subscription should return true",
			status:   stripe.SubscriptionStatusTrialing,
			expected: true,
		},
		{
			name:     "canceled subscription should return false",
			status:   stripe.SubscriptionStatusCanceled,
			expected: false,
		},
		{
			name:     "paused subscription should return false",
			status:   stripe.SubscriptionStatusPaused,
			expected: false,
		},
		{
			name:     "incomplete subscription should return false",
			status:   stripe.SubscriptionStatusIncomplete,
			expected: false,
		},
		{
			name:     "past due subscription should return false",
			status:   stripe.SubscriptionStatusPastDue,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSubscriptionActiveOrTrialing(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldUpdateCustomerPersonalOrgMetadata(t *testing.T) {
	tests := []struct {
		name     string
		customer *stripe.Customer
		expected bool
	}{
		{
			name: "customer without metadata should be updated",
			customer: &stripe.Customer{
				ID:       "cus_no_metadata",
				Metadata: nil,
			},
			expected: true,
		},
		{
			name: "customer without personal_org metadata should be updated",
			customer: &stripe.Customer{
				ID: "cus_missing_personal_org",
				Metadata: map[string]string{
					"some_other_field": "value",
				},
			},
			expected: true,
		},
		{
			name: "customer with personal_org=false should be updated",
			customer: &stripe.Customer{
				ID: "cus_personal_org_false",
				Metadata: map[string]string{
					"personal_org": "false",
				},
			},
			expected: true,
		},
		{
			name: "customer with personal_org=true should not be updated",
			customer: &stripe.Customer{
				ID: "cus_personal_org_true",
				Metadata: map[string]string{
					"personal_org": "true",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldUpdateCustomerPersonalOrgMetadata(tt.customer)
			assert.Equal(t, tt.expected, result)
		})
	}
}
