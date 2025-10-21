package hooks

import (
	"context"
	"reflect"
	"strings"

	"github.com/rs/zerolog"

	entgen "github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/file"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	"github.com/theopenlane/core/pkg/events/soiree"
)

const fileDetachedTopic = "file.detached"

func registerFileListeners(e *Eventer) error {
	_, err := e.Emitter.On(fileDetachedTopic, handleFileDetached)
	if err != nil {
		return err
	}

	return nil
}

func handleFileDetached(event soiree.Event) error {
	ctx := event.Context()

	idsVal := event.Properties().GetKey("file_ids")
	var fileIDs []string

	// Event properties are stored as generic slices; normalize to []string.
	switch v := idsVal.(type) {
	case []string:
		fileIDs = append(fileIDs, v...)
	case []any:
		for _, item := range v {
			if str, ok := item.(string); ok {
				fileIDs = append(fileIDs, str)
			}
		}
	}

	if len(fileIDs) == 0 {
		return nil
	}

	client, ok := event.Client().(*entgen.Client)
	if !ok || client == nil {
		return nil
	}

	// Use an allow-context so the orphan check and delete bypass privacy filters.
	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	for _, id := range fileIDs {
		orphaned, err := fileIsOrphan(allowCtx, client, id)
		if err != nil {
			zerolog.Ctx(ctx).Warn().Err(err).Str("file_id", id).Msg("unable to verify orphaned file")
			continue
		}
		if !orphaned {
			continue
		}

		if err := client.File.DeleteOneID(id).Exec(allowCtx); err != nil {
			zerolog.Ctx(ctx).Error().Err(err).Str("file_id", id).Msg("failed to delete orphaned file")
		}
	}

	return nil
}

func fileIsOrphan(ctx context.Context, client *entgen.Client, id string) (bool, error) {
	base := client.File.Query().Where(file.ID(id))
	baseType := reflect.TypeOf(base)

	for i := 0; i < baseType.NumMethod(); i++ {
		method := baseType.Method(i)
		if !strings.HasPrefix(method.Name, "Query") || len(method.Name) <= len("Query") {
			continue
		}

		// Clone the base query so each iteration starts from `WHERE id = ?`.
		clone := base.Clone()
		results := method.Func.Call([]reflect.Value{reflect.ValueOf(clone)})
		if len(results) != 1 {
			continue
		}

		subQuery := results[0]
		existMethod := subQuery.MethodByName("Exist")
		if !existMethod.IsValid() {
			continue
		}

		existResults := existMethod.Call([]reflect.Value{reflect.ValueOf(ctx)})
		if len(existResults) != 2 {
			continue
		}

		if errVal := existResults[1].Interface(); errVal != nil {
			return false, errVal.(error)
		}

		if existResults[0].Bool() {
			// Any edge still present means the file is referenced somewhere.
			return false, nil
		}
	}

	return true, nil
}

func emitFileDetachedEvent(e *Eventer, mutation entgen.Mutation, base soiree.Event, mutations []ClearedMutation) {
	fileIDs := fileIDsFromMutations(mutations)
	if len(fileIDs) == 0 {
		return
	}

	event := soiree.NewBaseEvent(fileDetachedTopic, mutation)
	event.SetContext(base.Context())
	event.SetClient(base.Client())

	props := soiree.NewProperties().
		Set("mutation_type", mutation.Type()).
		Set("operation", mutation.Op().String()).
		Set("file_ids", append([]string(nil), fileIDs...))
	event.SetProperties(props)

	e.Emitter.Emit(fileDetachedTopic, event)
}
