package entityops

import (
	"context"
	"encoding/json"
	"testing"

	generated "github.com/theopenlane/core/v2/internal/ent/generated"
)

func TestIngestNoopMark(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	if IngestNoopMarked(ctx) {
		t.Fatal("an unmarked context must not report a no-op")
	}

	markIngestNoop(ctx)

	ctx = WithIngestNoopMark(ctx)
	if IngestNoopMarked(ctx) {
		t.Fatal("a freshly installed holder must not report a no-op")
	}

	markIngestNoop(ctx)

	if !IngestNoopMarked(ctx) {
		t.Fatal("a marked holder must report a no-op")
	}

	if IngestNoopMarked(WithIngestNoopMark(ctx)) {
		t.Fatal("a fresh holder must shadow the outer mark")
	}
}

func TestSchemaUnchangedCapability(t *testing.T) {
	t.Parallel()

	if SchemaAsset.unchanged == nil {
		t.Fatal("SchemaAsset must carry the unchanged capability")
	}

	row, err := json.Marshal(&generated.Asset{Name: "Server"})
	if err != nil {
		t.Fatal(err)
	}

	same, err := SchemaAsset.unchanged(row, json.RawMessage(`{"name":"Server"}`))
	if err != nil || !same {
		t.Fatalf("identical payload must report unchanged, got same=%v err=%v", same, err)
	}

	same, err = SchemaAsset.unchanged(row, json.RawMessage(`{"name":"Renamed"}`))
	if err != nil || same {
		t.Fatalf("changed payload must report changed, got same=%v err=%v", same, err)
	}
}

func TestPayloadCarriesEdgeKeys(t *testing.T) {
	t.Parallel()

	s := &Schema{Edges: []EdgeDescriptor{
		{Name: "controls", CreateField: "control_ids", AddField: "add_control_ids"},
		{Name: "owner", Unique: true, CreateField: "owner_id", Field: "owner_id"},
		{Name: "parent", Unique: true, CreateField: "parent_id"},
	}}

	if s.payloadCarriesEdgeKeys(json.RawMessage(`{"name":"x"}`)) {
		t.Fatal("a payload without edge keys must not report edge keys")
	}

	if !s.payloadCarriesEdgeKeys(json.RawMessage(`{"control_ids":["c1"]}`)) {
		t.Fatal("a payload with a to-many create edge key must report edge keys")
	}

	if !s.payloadCarriesEdgeKeys(json.RawMessage(`{"add_control_ids":["c1"]}`)) {
		t.Fatal("a payload with a to-many add edge key must report edge keys")
	}

	if s.payloadCarriesEdgeKeys(json.RawMessage(`{"owner_id":"o1"}`)) {
		t.Fatal("a foreign-key-backed unique edge key is compared as a column and must not report edge keys")
	}

	if !s.payloadCarriesEdgeKeys(json.RawMessage(`{"parent_id":"p1"}`)) {
		t.Fatal("an edge-only unique key is invisible to the comparison and must report edge keys")
	}

	if !s.payloadCarriesEdgeKeys(json.RawMessage(`not json`)) {
		t.Fatal("an undecodable payload must conservatively report edge keys")
	}
}
