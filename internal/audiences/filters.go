package audiences

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/contact"
	"github.com/theopenlane/core/v2/internal/ent/generated/identityholder"
	"github.com/theopenlane/core/v2/pkg/celx"
)

const resolveBatchSize = 100

var (
	errAudienceFiltersRequired       = errors.New("audience filters must include at least one selector")
	errManualAudienceFilters         = errors.New("manual audiences cannot define filters")
	errDynamicAudienceFiltersMissing = errors.New("dynamic audiences require filters")
	errUnsupportedAudienceType       = errors.New("unsupported audience type")
	errSelectorSchemaRequired        = errors.New("selector schema is required")
	errSchemaNotRegistered           = errors.New("schema is not registered")
	errUnsupportedRecipientSource    = errors.New("schema cannot be used as an audience recipient source")
	errKeyMatchUnsupported           = errors.New("key_match is not supported for audience selectors yet")
	errSourceSelectorsUnsupported    = errors.New("source selectors are not supported for audience selectors yet")
	errSelectorSchemaNoExpressions   = errors.New("selector schema does not support expressions")
)

type filterSet struct {
	Selectors []entityops.TargetSelector `json:"selectors,omitempty"`
}

type Recipient struct {
	Email          string
	FullName       string
	ContactID      string
	UserID         string
	GroupID        string
	SubscriberID   string
	Source         string
	SourceObjectID string
	Metadata       map[string]any
}

type selectorResolveOptions[T any] struct {
	targetType reflect.Type
	fetchFn    func(lastKnownID string) ([]T, error)
	id         func(T) string
	recipient  func(T) Recipient
}

func parseSelectors(filters map[string]any) ([]entityops.TargetSelector, error) {
	if len(filters) == 0 {
		return nil, errAudienceFiltersRequired
	}

	buf, err := json.Marshal(filters)
	if err != nil {
		return nil, fmt.Errorf("marshal audience filters: %w", err)
	}

	if _, ok := filters["selectors"]; ok {
		var set filterSet
		if err := json.Unmarshal(buf, &set); err != nil {
			return nil, fmt.Errorf("decode audience selectors: %w", err)
		}

		return set.Selectors, nil
	}

	var selector entityops.TargetSelector
	if err := json.Unmarshal(buf, &selector); err != nil {
		return nil, fmt.Errorf("decode audience selector: %w", err)
	}

	return []entityops.TargetSelector{selector}, nil
}

func ValidateAudienceFilters(audienceType enums.AudienceType, filters map[string]any) error {
	switch audienceType {
	case enums.AudienceTypeManual:
		if len(filters) > 0 {
			return errManualAudienceFilters
		}

		return nil
	case enums.AudienceTypeDynamic:
		if len(filters) == 0 {
			return errDynamicAudienceFiltersMissing
		}

		return validateFilters(filters)
	default:
		return fmt.Errorf("%w: %q", errUnsupportedAudienceType, audienceType)
	}
}

func validateFilters(filters map[string]any) error {
	selectors, err := parseSelectors(filters)
	if err != nil {
		return err
	}

	if len(selectors) == 0 {
		return errAudienceFiltersRequired
	}

	for i, selector := range selectors {
		if err := validateSelector(selector); err != nil {
			return fmt.Errorf("selector %d: %w", i, err)
		}
	}

	return nil
}

func ResolveRecipients(ctx context.Context, db *generated.Client, orgID string, filters map[string]any, handle func([]Recipient) error) error {
	selectors, err := parseSelectors(filters)
	if err != nil {
		return err
	}

	for _, selector := range selectors {
		if err := validateSelector(selector); err != nil {
			return err
		}

		if err := resolveSelectors(ctx, db, orgID, selector, handle); err != nil {
			return err
		}
	}

	return nil
}

func validateSelector(selector entityops.TargetSelector) error {
	if selector.Schema.IsZero() {
		return errSelectorSchemaRequired
	}

	schema, ok := entityops.LookupSchema(selector.Schema.Name)
	if !ok {
		return fmt.Errorf("%w: %q", errSchemaNotRegistered, selector.Schema.Name)
	}

	switch schema.Snake {
	case entityops.SchemaContact.Snake, entityops.SchemaIdentityHolder.Snake:
	default:
		return fmt.Errorf("%w: %q", errUnsupportedRecipientSource, schema.Snake)
	}

	if selector.KeyMatch != nil {
		return errKeyMatchUnsupported
	}

	if len(selector.SourceContext) > 0 || !selector.SourceSchema.IsZero() {
		return errSourceSelectorsUnsupported
	}

	return nil
}

func resolveSelectors(ctx context.Context, db *generated.Client, orgID string, selector entityops.TargetSelector, handle func([]Recipient) error) error {
	schema, _ := entityops.LookupSchema(selector.Schema.Name)

	switch schema.Snake {
	case entityops.SchemaContact.Snake:
		return resolveSelector(ctx, selector, selectorResolveOptions[*generated.Contact]{
			targetType: reflect.TypeFor[entityops.ContactProjection](),
			fetchFn: func(lastKnownID string) ([]*generated.Contact, error) {
				query := db.Contact.Query().
					Where(
						contact.OwnerIDEQ(orgID),
						contact.EmailNotNil(),
						contact.EmailNEQ(""),
					).
					Order(contact.ByID()).
					Limit(resolveBatchSize)

				if lastKnownID != "" {
					query.Where(contact.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},
			id: func(c *generated.Contact) string {
				return c.ID
			},
			recipient: func(c *generated.Contact) Recipient {
				return Recipient{
					Email:          c.Email,
					FullName:       c.FullName,
					ContactID:      c.ID,
					Source:         entityops.SchemaContact.Snake,
					SourceObjectID: c.ID,
				}
			},
		}, handle)
	case entityops.SchemaIdentityHolder.Snake:
		return resolveSelector(ctx, selector, selectorResolveOptions[*generated.IdentityHolder]{
			targetType: reflect.TypeFor[entityops.IdentityHolderProjection](),
			fetchFn: func(lastKnownID string) ([]*generated.IdentityHolder, error) {
				query := db.IdentityHolder.Query().
					Where(
						identityholder.OwnerIDEQ(orgID),
						identityholder.EmailNEQ(""),
					).
					Order(identityholder.ByID()).
					Limit(resolveBatchSize)

				if lastKnownID != "" {
					query.Where(identityholder.IDGT(lastKnownID))
				}

				return query.All(ctx)
			},
			id: func(holder *generated.IdentityHolder) string {
				return holder.ID
			},
			recipient: func(holder *generated.IdentityHolder) Recipient {
				return Recipient{
					Email:          holder.Email,
					FullName:       holder.FullName,
					UserID:         holder.UserID,
					Source:         entityops.SchemaIdentityHolder.Snake,
					SourceObjectID: holder.ID,
				}
			},
		}, handle)
	default:
		return fmt.Errorf("%w: %q", errUnsupportedRecipientSource, schema.Snake)
	}
}

func resolveSelector[T any](ctx context.Context, selector entityops.TargetSelector, opts selectorResolveOptions[T], handle func([]Recipient) error) error {
	eval, err := buildCelEvaluator(opts.targetType)
	if err != nil {
		return err
	}

	var lastID string
	for {
		items, err := opts.fetchFn(lastID)
		if err != nil {
			return err
		}

		recipients := make([]Recipient, 0, len(items))
		for _, item := range items {
			lastID = opts.id(item)

			match, err := doesSelectorMatch(ctx, eval, selector.Expression, item)
			if err != nil {
				return err
			}

			if !match {
				continue
			}

			recipients = append(recipients, opts.recipient(item))
		}

		if len(recipients) > 0 {
			if err := handle(recipients); err != nil {
				return err
			}
		}

		if len(items) < resolveBatchSize {
			break
		}
	}

	return nil
}

func buildCelEvaluator(targetType reflect.Type) (*celx.NativeEntityEvaluator, error) {
	if targetType == nil {
		return nil, errSelectorSchemaNoExpressions
	}

	envCfg := celx.StrictEnvConfig()
	envCfg.CrossTypeNumericComparisons = true

	eval, err := celx.NewNativeEntityEvaluator(envCfg, celx.FastEvalConfig(), targetType, nil)
	if err != nil {
		return nil, fmt.Errorf("build selector evaluator: %w", err)
	}

	return eval, nil
}

func doesSelectorMatch(ctx context.Context, eval *celx.NativeEntityEvaluator, expression string, entity any) (bool, error) {
	if strings.TrimSpace(expression) == "" {
		return true, nil
	}

	data, err := json.Marshal(entity)
	if err != nil {
		return false, fmt.Errorf("marshal selector entity: %w", err)
	}

	match, err := eval.EvaluateBool(ctx, expression, data)
	if err != nil {
		return false, fmt.Errorf("evaluate selector expression: %w", err)
	}

	return match, nil
}
