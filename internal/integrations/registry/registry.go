package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/samber/lo"

	"github.com/theopenlane/core/v2/internal/ent/entityops"
	"github.com/theopenlane/core/v2/internal/integrations/types"
	"github.com/theopenlane/core/v2/pkg/gala"
	"github.com/theopenlane/core/v2/pkg/mapx"
)

// Builder builds one manifest-backed definition
type Builder func() (types.Definition, error)

// Registry is the in-memory index of registered definitions
type Registry struct {
	// definitions maps definition ID to its compiled entry
	definitions map[string]definitionEntry
	// operationsByTopic maps a topic name to its operation registration
	operationsByTopic map[gala.TopicName]types.OperationRegistration
	// webhookEventsByTopic maps a topic name to its webhook event registration
	webhookEventsByTopic map[gala.TopicName]types.WebhookEventRegistration
	// galaListeners collects standalone gala listener registrations across definitions
	galaListeners []types.GalaListenerRegistration
}

// Surface is the installation-facing surface of a definition: every kind that binds stored installation data to a definition type or name
type Surface struct {
	// ID is the canonical definition identifier
	ID string `json:"id"`
	// Credentials lists every credential slot with its stored schema, sorted by ref
	Credentials []SurfaceCredential `json:"credentials"`
	// UserInput is the stored user input schema when the definition declares a typed user input
	UserInput *SurfaceSchema `json:"userInput,omitempty"`
	// Installation is the derived installation metadata schema when the definition declares one
	Installation *SurfaceSchema `json:"installation,omitempty"`
	// Connections lists every connection mode's selecting credential ref, sorted
	Connections []string `json:"connections"`
	// Operations lists every operation name with the retired names it takes over and its config schema, sorted by name
	Operations []SurfaceOperation `json:"operations,omitempty"`
	// Webhooks lists every webhook contract with its events, sorted by name
	Webhooks []SurfaceWebhook `json:"webhooks,omitempty"`
}

// SurfaceSchema is the stored schema of one kind and how its stored payloads are carried across versions
type SurfaceSchema struct {
	// Schema is the reflected JSON schema of the stored type
	Schema json.RawMessage `json:"schema"`
	// Replaces lists the retired names whose stored payloads this kind takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
	// Backfill reports whether the kind declares a backfill for payloads missing values
	Backfill bool `json:"backfill,omitempty"`
}

// SurfaceCredential is one credential slot and the schema of what it stores
type SurfaceCredential struct {
	// Ref is the stable credential slot name
	Ref string `json:"ref"`
	// SurfaceSchema is the stored credential schema with its replacement and backfill declarations
	SurfaceSchema
}

// SurfaceNamed is one named registration and the retired names whose stored data it takes over
type SurfaceNamed struct {
	// Name is the stable registration name
	Name string `json:"name"`
	// Replaces lists the retired names this registration takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
}

// SurfaceOperation is one operation registration, the retired names it takes over, and the schema of its config
type SurfaceOperation struct {
	// Name is the stable operation name
	Name string `json:"name"`
	// Replaces lists the retired operation names this operation takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
	// Schema is the reflected JSON schema of the operation's config
	Schema json.RawMessage `json:"schema,omitempty"`
	// Section reports whether the operation's config is resolved from the stored user input rather than supplied by each caller
	Section bool `json:"section,omitempty"`
}

// SurfaceWebhook is one webhook contract with its events
type SurfaceWebhook struct {
	// Name is the stable webhook contract name
	Name string `json:"name"`
	// Replaces lists the retired contract names whose persisted webhook rows this contract takes over, sorted
	Replaces []string `json:"replaces,omitempty"`
	// Events lists the contract's events, sorted by name
	Events []SurfaceNamed `json:"events,omitempty"`
}

// DefinitionSurface projects a definition onto its installation-facing surface, sorted for stable encoding
func DefinitionSurface(def types.Definition) Surface {
	credentials := lo.Map(def.CredentialRegistrations, func(registration types.CredentialRegistration, _ int) SurfaceCredential {
		replaces := lo.Map(registration.Replaces, func(slot types.CredentialSlotID, _ int) string { return slot.String() })

		return SurfaceCredential{Ref: registration.Ref.String(), SurfaceSchema: SurfaceSchema{Schema: registration.StoredSchema, Replaces: sortedNames(replaces), Backfill: registration.Backfill != nil}}
	})

	slices.SortFunc(credentials, func(a, b SurfaceCredential) int {
		return strings.Compare(a.Ref, b.Ref)
	})

	connections := lo.Map(def.Connections, func(connection types.ConnectionRegistration, _ int) string {
		return connection.CredentialRef.String()
	})

	slices.Sort(connections)

	operations := sortedOperations(lo.Map(def.Operations, func(operation types.OperationRegistration, _ int) SurfaceOperation {
		return SurfaceOperation{Name: operation.Name, Replaces: sortedNames(operation.Replaces), Schema: operation.ConfigSchema, Section: operation.ConfigResolver != nil}
	}))

	webhooks := lo.Map(def.Webhooks, func(webhook types.WebhookRegistration, _ int) SurfaceWebhook {
		return SurfaceWebhook{Name: webhook.Name, Replaces: sortedNames(webhook.Replaces), Events: sortedNamed(lo.Map(webhook.Events, func(event types.WebhookEventRegistration, _ int) SurfaceNamed {
			return SurfaceNamed{Name: event.Name}
		}))}
	})

	slices.SortFunc(webhooks, func(a, b SurfaceWebhook) int {
		return strings.Compare(a.Name, b.Name)
	})

	surface := Surface{ID: def.ID, Credentials: credentials, Connections: connections, Operations: operations, Webhooks: webhooks}

	if def.UserInput != nil {
		surface.UserInput = &SurfaceSchema{Schema: def.UserInput.Schema, Replaces: sortedNames(def.UserInput.Replaces), Backfill: def.UserInput.Backfill != nil}
	}

	if def.Installation != nil && len(def.Installation.Schema) > 0 {
		surface.Installation = &SurfaceSchema{Schema: def.Installation.Schema}
	}

	return surface
}

// sortedNames returns a sorted copy of the names, nil when there are none
func sortedNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}

	return slices.Sorted(slices.Values(names))
}

// sortedNamed sorts named surface entries by name
func sortedNamed(entries []SurfaceNamed) []SurfaceNamed {
	slices.SortFunc(entries, func(a, b SurfaceNamed) int {
		return strings.Compare(a.Name, b.Name)
	})

	return entries
}

// sortedOperations sorts operation surface entries by name
func sortedOperations(entries []SurfaceOperation) []SurfaceOperation {
	slices.SortFunc(entries, func(a, b SurfaceOperation) int {
		return strings.Compare(a.Name, b.Name)
	})

	return entries
}

// definitionEntry captures the indexed details for one registered definition
type definitionEntry struct {
	// definition holds the original definition as supplied by the caller
	definition types.Definition
	// connections maps credential ref name to its connection registration
	connections map[string]types.ConnectionRegistration
	// clients maps client ID to its client registration
	clients map[types.ClientID]types.ClientRegistration
	// operations maps operation name to its operation registration
	operations map[string]types.OperationRegistration
	// webhooks maps webhook name to its webhook registration
	webhooks map[string]types.WebhookRegistration
	// webhookEvents maps webhook name to a nested map of event name to event registration
	webhookEvents map[string]map[string]types.WebhookEventRegistration
	// runtimeClient holds the pre-built client for runtime integrations
	runtimeClient any
	// version is the hash of the definition's declarative composition
	version string
}

// New constructs an empty registry
func New() *Registry {
	return &Registry{
		definitions:          map[string]definitionEntry{},
		operationsByTopic:    map[gala.TopicName]types.OperationRegistration{},
		webhookEventsByTopic: map[gala.TopicName]types.WebhookEventRegistration{},
	}
}

// Register adds one definition to the registry
func (r *Registry) Register(def types.Definition) error {
	def, err := finalizeDefinition(def)
	if err != nil {
		return err
	}

	if err := r.validateDefinition(def); err != nil {
		return err
	}

	populateMappingLinkTargets(def.Mappings)

	entry, err := compileDefinition(def)
	if err != nil {
		return err
	}

	if def.RuntimeIntegration != nil && def.RuntimeIntegration.Config != nil {
		client, buildErr := def.RuntimeIntegration.Build(context.Background(), def.RuntimeIntegration.Config)
		if buildErr != nil {
			return buildErr
		}

		entry.runtimeClient = client
	}

	operationTopics := make(map[gala.TopicName]types.OperationRegistration, len(def.Operations))

	for _, operation := range def.Operations {
		if lo.HasKey(r.operationsByTopic, operation.Topic) || lo.HasKey(operationTopics, operation.Topic) {
			return duplicateRegistration(def.ID, "operation topic", string(operation.Topic))
		}

		operationTopics[operation.Topic] = operation
	}

	webhookEventTopics := make(map[gala.TopicName]types.WebhookEventRegistration)

	for _, webhook := range def.Webhooks {
		for _, event := range webhook.Events {
			if lo.HasKey(r.webhookEventsByTopic, event.Topic) || lo.HasKey(webhookEventTopics, event.Topic) {
				return duplicateRegistration(def.ID, "webhook event topic", string(event.Topic))
			}

			webhookEventTopics[event.Topic] = event
		}
	}

	maps.Copy(r.operationsByTopic, operationTopics)
	maps.Copy(r.webhookEventsByTopic, webhookEventTopics)

	r.definitions[def.ID] = entry

	r.galaListeners = append(r.galaListeners, def.GalaListeners...)

	return nil
}

// duplicateRegistration reports a second registration of one kind under a name a definition already holds
func duplicateRegistration(definitionID, kind, name string) error {
	return fmt.Errorf("%w: definition %s %s %q", ErrDuplicateRegistration, definitionID, kind, name)
}

// RegisterAll builds and registers every supplied definition builder in order
func (r *Registry) RegisterAll(builders ...Builder) error {
	for _, builder := range builders {
		if builder == nil {
			return ErrBuilderNil
		}

		def, err := builder()
		if err != nil {
			return err
		}

		if err := r.Register(def); err != nil {
			return err
		}
	}

	return nil
}

// Definition returns one definition by canonical identifier
func (r *Registry) Definition(id string) (types.Definition, bool) {
	entry, ok := r.definitions[id]
	if !ok {
		return types.Definition{}, false
	}

	return entry.definition, true
}

// Version returns the computed version of one definition, or empty when unregistered
func (r *Registry) Version(id string) string {
	entry, ok := r.definitions[id]
	if !ok {
		return ""
	}

	return entry.version
}

// Definitions returns all registered definitions in stable id order
func (r *Registry) Definitions() []types.Definition {
	return mapx.SortedProjection(r.definitions, func(e definitionEntry) types.Definition { return e.definition }, func(d types.Definition) string { return d.ID })
}

// Client returns one client registration for a definition
func (r *Registry) Client(id string, clientID types.ClientID) (types.ClientRegistration, error) {
	return lookupInEntry(r, id, clientID, func(e definitionEntry) map[types.ClientID]types.ClientRegistration {
		return e.clients
	}, ErrClientNotFound)
}

// Operation returns one operation registration for a definition
func (r *Registry) Operation(id string, name string) (types.OperationRegistration, error) {
	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.OperationRegistration {
		return e.operations
	}, ErrOperationNotFound)
}

// Webhook returns one webhook registration for a definition
func (r *Registry) Webhook(id string, name string) (types.WebhookRegistration, error) {
	return lookupInEntry(r, id, name, func(e definitionEntry) map[string]types.WebhookRegistration {
		return e.webhooks
	}, ErrWebhookNotFound)
}

// Catalog returns all definition specs in stable id order
func (r *Registry) Catalog() []types.DefinitionSpec {
	return mapx.SortedProjection(r.definitions, func(e definitionEntry) types.DefinitionSpec { return e.definition.DefinitionSpec }, func(s types.DefinitionSpec) string { return s.ID })
}

// validateDefinition checks the top-level definition identity fields before registration
func (r *Registry) validateDefinition(def types.Definition) error {
	if def.ID == "" {
		return ErrDefinitionIDRequired
	}

	if _, exists := r.definitions[def.ID]; exists {
		return ErrDefinitionAlreadyRegistered
	}

	if def.OperatorConfig != nil && len(def.OperatorConfig.Schema) == 0 {
		return ErrOperatorConfigSchemaRequired
	}

	if lo.ContainsBy(def.CredentialRegistrations, func(credential types.CredentialRegistration) bool {
		return len(credential.StoredSchema) == 0
	}) {
		return ErrCredentialSchemaRequired
	}

	if def.UserInput != nil && len(def.UserInput.Schema) == 0 {
		return ErrUserInputSchemaRequired
	}

	if err := validateHealthCheck(def); err != nil {
		return err
	}

	if def.RuntimeIntegration != nil {
		if def.RuntimeIntegration.Build == nil {
			return ErrRuntimeBuildRequired
		}
	}

	if err := validateMappingLinks(def.Mappings); err != nil {
		return err
	}

	return nil
}

// validateHealthCheck requires a health check with a handler when the definition declares connections, and a health check client built from every connection's credential slot
func validateHealthCheck(def types.Definition) error {
	check := def.HealthCheck

	switch {
	case check == nil && len(def.Connections) > 0:
		return ErrHealthCheckRequired
	case check == nil:
		return nil
	case check.Handle == nil:
		return ErrConnectionHealthCheckHandlerRequired
	case !check.ClientRef.Valid():
		return nil
	}

	client, declared := lo.Find(def.Clients, func(client types.ClientRegistration) bool {
		return client.Ref == check.ClientRef
	})
	if !declared {
		return ErrConnectionClientRefNotDeclared
	}

	uncovered, found := lo.Find(def.Connections, func(connection types.ConnectionRegistration) bool {
		return !lo.Contains(client.CredentialRefs, connection.CredentialRef)
	})
	if found {
		return fmt.Errorf("%w: definition %s client %s connection %s", ErrHealthCheckClientCredentialMissing, def.ID, check.ClientRef, uncovered.CredentialRef)
	}

	return nil
}

// populateMappingLinkTargets fills each mapping's cross-link inventory from the entityops catalog — one entry per edge, carrying the edge name, the target's match-key fields, and the source's mapped input keys — so the definition payload surfaces the exact identifiers a LinkRule may reference and configuration never falls back to free-typed field names
func populateMappingLinkTargets(mappings []types.MappingRegistration) {
	for i := range mappings {
		sourceSchema, ok := entityops.LookupSchema(mappings[i].Schema)
		if !ok {
			continue
		}

		sourceFields := sourceLinkFields(sourceSchema.Fields)

		var targets []types.LinkTargetInfo

		for _, edge := range sourceSchema.Edges {
			if edge.TargetType == "" || edge.Target == nil {
				continue
			}

			targets = append(targets, types.LinkTargetInfo{
				Edge:         edge.Name,
				TargetType:   edge.TargetType,
				Label:        edge.Label,
				TargetFields: targetLinkFields(edge.Target.Fields),
				SourceFields: sourceFields,
			})
		}

		mappings[i].LinkTargets = targets
	}
}

// targetLinkFields projects the match-key (indexed string) fields of a target schema — the fields a LinkRule.TargetField may name
func targetLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if !f.MatchKey {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.Name, Label: f.Label, Type: f.Type}, true
	})
}

// sourceLinkFields projects the mapped input keys of the source schema — the keys present in the mapped ingest payload that a LinkRule.SourceField (scalar) or SourceList (list) may name
func sourceLinkFields(fields []entityops.FieldDescriptor) []types.LinkFieldInfo {
	return lo.FilterMap(fields, func(f entityops.FieldDescriptor, _ int) (types.LinkFieldInfo, bool) {
		if f.InputKey == "" {
			return types.LinkFieldInfo{}, false
		}

		return types.LinkFieldInfo{Name: f.InputKey, Label: f.Label, Type: f.Type}, true
	})
}

// ResolveLinkEdge resolves the edge a link rule links through: an explicit rule edge is looked up by name and checked against the declared target type; otherwise the target type must identify exactly one edge, since silently picking one of several (e.g. editors vs viewers, both targeting Group) would link through an arbitrary edge
func ResolveLinkEdge(sourceSchema *entityops.Schema, rule types.LinkRule) (entityops.EdgeDescriptor, error) {
	if rule.Edge != "" {
		edge, found := sourceSchema.EdgeByName(rule.Edge)
		if !found {
			return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has no edge %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.Edge)
		}

		if rule.TargetSchema != "" && edge.TargetType != rule.TargetSchema {
			return entityops.EdgeDescriptor{}, fmt.Errorf("%w: edge %s.%s targets %s, not %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.Edge, edge.TargetType, rule.TargetSchema)
		}

		return edge, nil
	}

	candidates := lo.Filter(sourceSchema.Edges, func(e entityops.EdgeDescriptor, _ int) bool {
		return e.TargetType == rule.TargetSchema
	})

	switch len(candidates) {
	case 0:
		return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has no edge to %s", ErrLinkEdgeNotFound, sourceSchema.Name, rule.TargetSchema)
	case 1:
		return candidates[0], nil
	default:
		names := lo.Map(candidates, func(e entityops.EdgeDescriptor, _ int) string { return e.Name })

		return entityops.EdgeDescriptor{}, fmt.Errorf("%w: %s has %d edges to %s (%s); set the rule's edge", ErrLinkEdgeAmbiguous, sourceSchema.Name, len(candidates), rule.TargetSchema, strings.Join(names, ", "))
	}
}

// validateMappingLinks verifies every link rule a mapping declares against the entityops catalog — the edge resolves unambiguously, the match shape is coherent, the target field is a match key on the target, and the source fields are mapped input keys of the right shape — so a typo or an ambiguous target in a definition's link defaults fails at registration instead of silently misbehaving at ingest
func validateMappingLinks(mappings []types.MappingRegistration) error {
	for _, mapping := range mappings {
		if len(mapping.Spec.Links) == 0 {
			continue
		}

		sourceSchema, ok := entityops.LookupSchema(mapping.Schema)
		if !ok {
			continue
		}

		if err := ValidateLinkRules(sourceSchema, mapping.Spec.Links); err != nil {
			return err
		}
	}

	return nil
}

// ValidateLinkRules validates each rule against the source schema's catalog: the edge resolves unambiguously, the match shape is coherent, and the referenced fields exist with the right shape
func ValidateLinkRules(sourceSchema *entityops.Schema, rules []types.LinkRule) error {
	for _, rule := range rules {
		edge, err := ResolveLinkEdge(sourceSchema, rule)
		if err != nil {
			return err
		}

		if err := validateLinkRuleFields(sourceSchema, edge, rule); err != nil {
			return err
		}
	}

	return nil
}

// validateLinkRuleFields checks one resolved rule's match configuration; edges targeting an unregistered schema are rejected since link resolution needs the target catalog
func validateLinkRuleFields(sourceSchema *entityops.Schema, edge entityops.EdgeDescriptor, rule types.LinkRule) error {
	if edge.Target == nil {
		return fmt.Errorf("%w: %s.%s targets %s", ErrLinkTargetNotRegistered, sourceSchema.Name, edge.Name, edge.TargetType)
	}

	fieldMatch := rule.TargetField != "" && (rule.SourceField != "" || rule.SourceList != "")

	if fieldMatch == (rule.Expression != "") {
		return fmt.Errorf("%w: %s -> %s must set either a target/source field match or an expression", ErrLinkRuleInvalid, sourceSchema.Name, edge.Name)
	}

	if !fieldMatch {
		return nil
	}

	if !edge.Target.MatchKeyField(rule.TargetField) {
		return fmt.Errorf("%w: %s is not a match-key field on %s", ErrLinkTargetFieldInvalid, rule.TargetField, edge.TargetType)
	}

	if rule.SourceField != "" {
		if err := validateSourceKey(sourceSchema, rule.SourceField, false); err != nil {
			return err
		}
	}

	if rule.SourceList != "" {
		if err := validateSourceKey(sourceSchema, rule.SourceList, true); err != nil {
			return err
		}
	}

	return nil
}

// validateSourceKey checks that key is a mapped input key on the source schema whose type shape (scalar vs list) matches its LinkRule slot
func validateSourceKey(sourceSchema *entityops.Schema, key string, wantList bool) error {
	field, found := lo.Find(sourceSchema.Fields, func(f entityops.FieldDescriptor) bool {
		return f.InputKey == key
	})
	if !found {
		return fmt.Errorf("%w: %s is not a mapped input key on %s", ErrLinkSourceFieldInvalid, key, sourceSchema.Name)
	}

	if isList := strings.HasPrefix(field.Type, "[]"); isList != wantList {
		slot := "sourceField"
		if wantList {
			slot = "sourceList"
		}

		return fmt.Errorf("%w: %s.%s has type %s, which does not fit %s", ErrLinkSourceFieldInvalid, sourceSchema.Name, key, field.Type, slot)
	}

	return nil
}

// compileDefinition builds the indexed client, operation, and webhook event maps for one definition
func compileDefinition(def types.Definition) (definitionEntry, error) {
	credentialNames := indexCredentialNames(def.CredentialRegistrations)

	clients, err := indexClients(def.ID, def.Clients, credentialNames)
	if err != nil {
		return definitionEntry{}, err
	}

	operations, err := indexOperations(def.ID, def.Operations, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	connections, err := indexConnections(def.ID, def.Connections, credentialNames, clients)
	if err != nil {
		return definitionEntry{}, err
	}

	webhooks, webhookEvents, err := indexWebhooks(def.ID, def.Webhooks)
	if err != nil {
		return definitionEntry{}, err
	}

	version, err := computeVersion(def)
	if err != nil {
		return definitionEntry{}, err
	}

	return definitionEntry{
		definition:    def,
		connections:   connections,
		clients:       clients,
		operations:    operations,
		webhooks:      webhooks,
		webhookEvents: webhookEvents,
		version:       version,
	}, nil
}

// indexCredentialNames builds a set of declared credential ref names for a definition
func indexCredentialNames(registrations []types.CredentialRegistration) map[string]struct{} {
	return lo.SliceToMap(registrations, func(reg types.CredentialRegistration) (string, struct{}) {
		return reg.Ref.String(), struct{}{}
	})
}

// indexClients indexes client registrations by client ref while validating credential cross-references and rejecting repeated refs
func indexClients(definitionID string, clients []types.ClientRegistration, credentialNames map[string]struct{}) (map[types.ClientID]types.ClientRegistration, error) {
	index := make(map[types.ClientID]types.ClientRegistration, len(clients))

	for _, client := range clients {
		if !client.Ref.Valid() {
			return nil, ErrClientRequired
		}

		for _, ref := range client.CredentialRefs {
			if _, declared := credentialNames[ref.String()]; !declared {
				return nil, ErrCredentialRefNotDeclared
			}
		}

		if lo.HasKey(index, client.Ref) {
			return nil, duplicateRegistration(definitionID, "client", client.Ref.String())
		}

		index[client.Ref] = client
	}

	return index, nil
}

// indexConnections indexes connection registrations while enforcing credential, client, auth, disconnect, and unique slot constraints
func indexConnections(definitionID string, connections []types.ConnectionRegistration, credentialNames map[string]struct{}, clients map[types.ClientID]types.ClientRegistration) (map[string]types.ConnectionRegistration, error) {
	connectionIndex := make(map[string]types.ConnectionRegistration, len(connections))

	for _, connection := range connections {
		if connection.CredentialRef == (types.CredentialSlotID{}) {
			return nil, ErrConnectionCredentialRefRequired
		}

		name := connection.CredentialRef.String()

		if _, declared := credentialNames[name]; !declared {
			return nil, ErrConnectionCredentialRefNotDeclared
		}

		if lo.HasKey(connectionIndex, name) {
			return nil, duplicateRegistration(definitionID, "connection", name)
		}

		if !lo.Contains(connection.CredentialRefs, connection.CredentialRef) {
			connection.CredentialRefs = append(connection.CredentialRefs, connection.CredentialRef)
		}

		for _, ref := range connection.CredentialRefs {
			if _, declared := credentialNames[ref.String()]; !declared {
				return nil, ErrConnectionCredentialRefNotDeclared
			}
		}

		for _, clientRef := range connection.ClientRefs {
			if _, declared := clients[clientRef]; !declared {
				return nil, ErrConnectionClientRefNotDeclared
			}
		}

		if connection.Auth != nil {
			if connection.Auth.CredentialRef == (types.CredentialSlotID{}) {
				return nil, ErrConnectionAuthCredentialRefNotDeclared
			}

			if !lo.Contains(connection.CredentialRefs, connection.Auth.CredentialRef) {
				return nil, ErrConnectionAuthCredentialRefNotDeclared
			}
		}

		if connection.Disconnect != nil {
			if connection.Disconnect.CredentialRef == (types.CredentialSlotID{}) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}

			if !lo.Contains(connection.CredentialRefs, connection.Disconnect.CredentialRef) {
				return nil, ErrConnectionDisconnectCredentialRefNotDeclared
			}
		}

		connectionIndex[name] = connection
	}

	return connectionIndex, nil
}

// indexOperations indexes operations by name while validating handler and client cross-references and rejecting repeated names
func indexOperations(definitionID string, operations []types.OperationRegistration, clients map[types.ClientID]types.ClientRegistration) (map[string]types.OperationRegistration, error) {
	index := make(map[string]types.OperationRegistration, len(operations))

	for _, operation := range operations {
		switch {
		case operation.Handle == nil && operation.IngestHandle == nil:
			return nil, ErrOperationHandlerRequired
		case operation.Handle != nil && operation.IngestHandle != nil:
			return nil, ErrOperationHandlerAmbiguous
		case operation.IngestHandle != nil && len(operation.Ingest) == 0:
			return nil, ErrIngestContractsRequired
		}

		if operation.Policy.Snapshot && operation.IngestHandle == nil {
			return nil, ErrIngestSnapshotRequiresIngestHandle
		}

		if operation.ClientRef.Valid() {
			if _, exists := clients[operation.ClientRef]; !exists {
				return nil, ErrClientNotFound
			}
		}

		if lo.HasKey(index, operation.Name) {
			return nil, duplicateRegistration(definitionID, "operation", operation.Name)
		}

		index[operation.Name] = operation
	}

	return index, nil
}

// indexWebhooks indexes webhook contracts and webhook events while validating structural constraints and rejecting repeated names
func indexWebhooks(definitionID string, webhooks []types.WebhookRegistration) (map[string]types.WebhookRegistration, map[string]map[string]types.WebhookEventRegistration, error) {
	webhookIndex := make(map[string]types.WebhookRegistration, len(webhooks))
	webhookEventIndex := make(map[string]map[string]types.WebhookEventRegistration, len(webhooks))

	for _, webhook := range webhooks {
		if len(webhook.Events) > 0 && webhook.Event == nil {
			return nil, nil, ErrWebhookEventResolverRequired
		}

		if lo.HasKey(webhookIndex, webhook.Name) {
			return nil, nil, duplicateRegistration(definitionID, "webhook", webhook.Name)
		}

		eventIndex := make(map[string]types.WebhookEventRegistration, len(webhook.Events))

		for _, event := range webhook.Events {
			if event.Handle == nil {
				return nil, nil, ErrWebhookEventHandlerRequired
			}

			if lo.HasKey(eventIndex, event.Name) {
				return nil, nil, duplicateRegistration(definitionID, "webhook event", event.Name)
			}

			eventIndex[event.Name] = event
		}

		webhookIndex[webhook.Name] = webhook
		webhookEventIndex[webhook.Name] = eventIndex
	}

	return webhookIndex, webhookEventIndex, nil
}

// Listeners returns all operation registrations in stable topic order
func (r *Registry) Listeners() []types.OperationRegistration {
	return mapx.SortedValues(r.operationsByTopic, func(o types.OperationRegistration) gala.TopicName { return o.Topic })
}

// WebhookEvent returns one webhook event registration for a definition
func (r *Registry) WebhookEvent(id string, webhookName string, eventName string) (types.WebhookEventRegistration, error) {
	entry, ok := r.definitions[id]
	if !ok {
		return types.WebhookEventRegistration{}, ErrDefinitionNotFound
	}

	eventIndex, ok := entry.webhookEvents[webhookName]
	if !ok {
		return types.WebhookEventRegistration{}, ErrWebhookNotFound
	}

	event, ok := eventIndex[eventName]
	if !ok {
		return types.WebhookEventRegistration{}, ErrWebhookNotFound
	}

	return event, nil
}

// WebhookListeners returns all webhook event registrations in stable topic order
func (r *Registry) WebhookListeners() []types.WebhookEventRegistration {
	return mapx.SortedValues(r.webhookEventsByTopic, func(e types.WebhookEventRegistration) gala.TopicName { return e.Topic })
}

// GalaListeners returns all registered standalone gala listener registrations
func (r *Registry) GalaListeners() []types.GalaListenerRegistration {
	return append([]types.GalaListenerRegistration(nil), r.galaListeners...)
}

// RuntimeClient returns the cached runtime client for the given definition ID
func (r *Registry) RuntimeClient(definitionID string) (any, bool) {
	entry, ok := r.definitions[definitionID]
	if !ok || entry.runtimeClient == nil {
		return nil, false
	}

	return entry.runtimeClient, true
}

// StaticWebhooks returns all webhook registrations that declare a fixed static route
func (r *Registry) StaticWebhooks() []types.StaticWebhookEntry {
	var entries []types.StaticWebhookEntry

	for defID, entry := range r.definitions {
		for _, webhook := range entry.definition.Webhooks {
			if webhook.StaticRoute != "" {
				entries = append(entries, types.StaticWebhookEntry{
					DefinitionID: defID,
					WebhookName:  webhook.Name,
					StaticRoute:  webhook.StaticRoute,
				})
			}
		}
	}

	return entries
}

// IsRuntimeIntegration reports whether the given definition was provisioned as a runtime integration (no DB record, no keystore)
func (r *Registry) IsRuntimeIntegration(definitionID string) bool {
	entry, ok := r.definitions[definitionID]

	return ok && entry.definition.RuntimeIntegration != nil
}

// lookupInEntry finds an entry by definition id, then looks up a value in the sub-map returned by getMap
func lookupInEntry[K comparable, V any](r *Registry, id string, key K, getMap func(definitionEntry) map[K]V, notFoundErr error) (V, error) {
	entry, ok := r.definitions[id]
	if !ok {
		var zero V
		return zero, ErrDefinitionNotFound
	}

	val, ok := getMap(entry)[key]
	if !ok {
		return val, notFoundErr
	}

	return val, nil
}
