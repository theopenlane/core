package graphapi

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ph "github.com/posthog/posthog-go"
	"github.com/theopenlane/utils/slack"

	ent "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/integration"
	"github.com/theopenlane/core/internal/ent/generated/webhook"
	"github.com/theopenlane/core/pkg/events/soiree"
)

// CreateEvent creates an event for the mutation with the properties
func CreateEvent(ctx context.Context, c *ent.Client, m ent.Mutation, v ent.Value) {
	pool := soiree.NewPondPool(100, 1000)
	e := soiree.NewEventPool(soiree.WithPool(pool))

	out, err := parseValue(v)
	if err != nil {
		return
	}

	obj := strings.ToLower(m.Type())
	action := getOp(m)

	// debug log the event
	c.Logger.Debugw("tracking event", "object", obj, "action", action)

	event := fmt.Sprintf("%s.%sd", obj, action)
	e.EnsureTopic(event)

	id, ok := out["id"]
	if !ok {
		// keep going
		return
	}

	i, ok := id.(string)
	if !ok {
		// keep going
		return
	}

	// Set properties for the event
	// all events will have the id
	props := ph.NewProperties().
		Set(fmt.Sprintf("%s_id", obj), i)

	payload := map[string]string{"key": "value"}
	sEvent := soiree.NewBaseEvent(event, payload)

	// set the name if it exists
	name, ok := out["name"]
	if ok {
		props.Set(fmt.Sprintf("%s_name", obj), name)
		payload["name"] = name.(string)
	}

	// set the first name if it exists
	fName, ok := out["first_name"]
	if ok {
		props.Set("first_name", fName)
		payload["first_name"] = fName.(string)
	}

	// set the last name if it exists
	lName, ok := out["last_name"]
	if ok {
		props.Set("last_name", lName)
		payload["last_name"] = lName.(string)
	}

	// set the email if it exists
	email, ok := out["email"]
	if ok {
		props.Set("email", email)
		payload["email"] = email.(string)
	}

	authprovider, ok := out["auth_provider"]
	if ok {
		props.Set("auth_provider", authprovider)
		payload["auth_provider"] = authprovider.(string)
	}

	userCreatedListener := userCreatedListener(ctx, c, sEvent)
	orgCreatedListener := orgCreatedListener(ctx, c, sEvent)

	_, err = e.On("user.created", userCreatedListener)
	if err != nil {
		return
	}

	_, err = e.On("organization.created", orgCreatedListener)
	if err != nil {
		return
	}

	e.Emit(event, payload)

	c.Analytics.Event(event, props)

	// debug log the event
	c.Logger.Debugw("event tracked", "event", event, "props", props)
}

// trackedEvent returns true if the mutation should be a tracked event
// for now, lets just track high level create and delete events
// TODO: make these configurable by integration
func TrackedEvent(m ent.Mutation) bool {
	switch m.Type() {
	case "User", "Organization", "Group", "Subscriber":
		switch getOp(m) {
		case ActionCreate, ActionDelete:
			return true
		}
		return false
	}

	return false
}

// getOp returns the string action for the mutation
func getOp(m ent.Mutation) string {
	switch m.Op() {
	case ent.OpCreate:
		return ActionCreate
	case ent.OpUpdate, ent.OpUpdateOne:
		return ActionUpdate
	case ent.OpDelete, ent.OpDeleteOne:
		return ActionDelete
	default:
		return ""
	}
}

// parseValue returns a map of the ent.Value
func parseValue(v ent.Value) (map[string]interface{}, error) {
	out, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var valMap map[string]interface{}

	if err := json.Unmarshal(out, &valMap); err != nil {
		return nil, err
	}

	return valMap, nil
}

// userCreatedListener is a listener for the user created event
func userCreatedListener(ctx context.Context, c *ent.Client, sEvent *soiree.BaseEvent) func(evt soiree.Event) error {
	return func(evt soiree.Event) error {
		integrationWithWebhook, err := c.Integration.Query().WithWebhooks().Where(
			integration.KindEQ("slack")).QueryWebhooks().Where(
			webhook.EnabledEQ(true)).All(ctx)
		if err != nil {
			return err
		}

		for _, w := range integrationWithWebhook {
			retrieve := sEvent.Payload().(map[string]string)

			payload := slack.Payload{
				Text: fmt.Sprintf("A user with the following details has been created:\nName: %s\nFirst Name: %s\nLast Name: %s\nEmail: %s\nAuth Provider: %s",
					retrieve["name"],
					retrieve["first_name"],
					retrieve["last_name"],
					retrieve["email"],
					retrieve["auth_provider"]),
			}

			slackMessage := slack.New(w.DestinationURL)
			if err := slackMessage.Post(context.Background(), &payload); err != nil {
				return err
			}

		}
		return nil
	}
}

// orgCreatedListener is a listener for the organization created event
func orgCreatedListener(ctx context.Context, c *ent.Client, sEvent *soiree.BaseEvent) func(evt soiree.Event) error {
	return func(evt soiree.Event) error {
		integrationWithWebhook, err := c.Integration.Query().WithWebhooks().Where(
			integration.KindEQ("slack")).QueryWebhooks().Where(
			webhook.EnabledEQ(true)).All(ctx)
		if err != nil {
			return err
		}

		for _, w := range integrationWithWebhook {
			retrieve := sEvent.Payload().(map[string]string)

			payload := slack.Payload{
				Text: fmt.Sprintf("A user with the following details has been created:\nName: %s\nFirst Name: %s\nLast Name: %s\nEmail: %s\nAuth Provider: %s",
					retrieve["name"],
					retrieve["first_name"],
					retrieve["last_name"],
					retrieve["email"],
					retrieve["auth_provider"]),
			}

			slackMessage := slack.New(w.DestinationURL)
			if err := slackMessage.Post(context.Background(), &payload); err != nil {
				return err
			}

		}
		return nil
	}
}
