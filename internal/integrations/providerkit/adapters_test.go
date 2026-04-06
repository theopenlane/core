package providerkit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

var errBadConfig = errors.New("bad config")

func TestWithClient(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		handler := WithClient[string, json.RawMessage](ref, func(ctx context.Context, client string) (json.RawMessage, error) {
			return json.RawMessage(`{"client":"` + client + `"}`), nil
		})

		got, err := handler(context.Background(), types.OperationRequest{Client: "my-client"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var m map[string]string
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m["client"] != "my-client" {
			t.Fatalf("expected client=my-client, got %q", m["client"])
		}
	})

	t.Run("cast failure", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		handler := WithClient[string, json.RawMessage](ref, func(ctx context.Context, client string) (json.RawMessage, error) {
			return nil, nil
		})

		_, err := handler(context.Background(), types.OperationRequest{Client: 42})
		if !errors.Is(err, types.ErrClientCastFailed) {
			t.Fatalf("expected ErrClientCastFailed, got %v", err)
		}
	})
}

func TestWithClientRequest(t *testing.T) {
	t.Parallel()

	t.Run("success with full request access", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		handler := WithClientRequest[string, json.RawMessage](ref, func(ctx context.Context, req types.OperationRequest, client string) (json.RawMessage, error) {
			if req.Config == nil {
				return nil, errors.New("expected config in request")
			}

			return json.RawMessage(`{"ok":true}`), nil
		})

		got, err := handler(context.Background(), types.OperationRequest{
			Client: "test-client",
			Config: json.RawMessage(`{"key":"value"}`),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var m map[string]bool
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if !m["ok"] {
			t.Fatal("expected ok=true")
		}
	})
}

func TestWithClientConfig(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		Limit int `json:"limit"`
	}

	t.Run("success with typed config", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		op := types.NewOperationRef[testConfig]("test-op")

		handler := WithClientConfig[string, testConfig, json.RawMessage](ref, op, errBadConfig, func(ctx context.Context, client string, cfg testConfig) (json.RawMessage, error) {
			if cfg.Limit != 10 {
				return nil, errors.New("expected limit=10")
			}

			return json.RawMessage(`{"limit":10}`), nil
		})

		got, err := handler(context.Background(), types.OperationRequest{
			Client: "test-client",
			Config: json.RawMessage(`{"limit":10}`),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var m map[string]float64
		if err := json.Unmarshal(got, &m); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}

		if m["limit"] != 10 {
			t.Fatalf("expected limit=10, got %v", m["limit"])
		}
	})

	t.Run("bad config returns custom configErr", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		op := types.NewOperationRef[testConfig]("test-op")

		handler := WithClientConfig[string, testConfig, json.RawMessage](ref, op, errBadConfig, func(ctx context.Context, client string, cfg testConfig) (json.RawMessage, error) {
			return nil, nil
		})

		_, err := handler(context.Background(), types.OperationRequest{
			Client: "test-client",
			Config: json.RawMessage(`not-json`),
		})
		if !errors.Is(err, errBadConfig) {
			t.Fatalf("expected errBadConfig, got %v", err)
		}
	})
}

func TestWithClientRequest_Ingest(t *testing.T) {
	t.Parallel()

	t.Run("success returns payload sets", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		handler := WithClientRequest[string, []types.IngestPayloadSet](ref, func(ctx context.Context, req types.OperationRequest, client string) ([]types.IngestPayloadSet, error) {
			return []types.IngestPayloadSet{
				{Schema: "vuln.v1", Envelopes: []types.MappingEnvelope{{Variant: "alert"}}},
			}, nil
		})

		got, err := handler(context.Background(), types.OperationRequest{Client: "test-client"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 {
			t.Fatalf("expected 1 payload set, got %d", len(got))
		}

		if got[0].Schema != "vuln.v1" {
			t.Fatalf("expected schema=vuln.v1, got %q", got[0].Schema)
		}
	})

	t.Run("cast failure", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		handler := WithClientRequest[string, []types.IngestPayloadSet](ref, func(ctx context.Context, req types.OperationRequest, client string) ([]types.IngestPayloadSet, error) {
			return nil, nil
		})

		_, err := handler(context.Background(), types.OperationRequest{Client: 42})
		if !errors.Is(err, types.ErrClientCastFailed) {
			t.Fatalf("expected ErrClientCastFailed, got %v", err)
		}
	})
}

func TestWithClientRequestConfig_Ingest(t *testing.T) {
	t.Parallel()

	type testConfig struct {
		MaxItems int `json:"maxItems"`
	}

	t.Run("success with config", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		op := types.NewOperationRef[testConfig]("ingest-op")

		handler := WithClientRequestConfig[string, testConfig, []types.IngestPayloadSet](ref, op, errBadConfig, func(ctx context.Context, req types.OperationRequest, client string, cfg testConfig) ([]types.IngestPayloadSet, error) {
			if cfg.MaxItems != 5 {
				return nil, errors.New("expected maxItems=5")
			}

			return []types.IngestPayloadSet{{Schema: "items.v1"}}, nil
		})

		got, err := handler(context.Background(), types.OperationRequest{
			Client: "test-client",
			Config: json.RawMessage(`{"maxItems":5}`),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(got) != 1 || got[0].Schema != "items.v1" {
			t.Fatalf("unexpected result: %+v", got)
		}
	})

	t.Run("bad config returns custom configErr", func(t *testing.T) {
		t.Parallel()

		ref := types.NewClientRef[string]()
		op := types.NewOperationRef[testConfig]("ingest-op")

		handler := WithClientRequestConfig[string, testConfig, []types.IngestPayloadSet](ref, op, errBadConfig, func(ctx context.Context, req types.OperationRequest, client string, cfg testConfig) ([]types.IngestPayloadSet, error) {
			return nil, nil
		})

		_, err := handler(context.Background(), types.OperationRequest{
			Client: "test-client",
			Config: json.RawMessage(`not-json`),
		})
		if !errors.Is(err, errBadConfig) {
			t.Fatalf("expected errBadConfig, got %v", err)
		}
	})
}
