package providerkit

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/theopenlane/core/internal/integrations/types"
)

func TestEvalFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		envelope types.MappingEnvelope
		want     bool
		wantErr  error
	}{
		{
			name:     "empty expression returns true",
			expr:     "",
			envelope: types.MappingEnvelope{},
			want:     true,
		},
		{
			name: "matching expression returns true",
			expr: `variant == "alert"`,
			envelope: types.MappingEnvelope{
				Variant: "alert",
			},
			want: true,
		},
		{
			name: "non-matching expression returns false",
			expr: `variant == "alert"`,
			envelope: types.MappingEnvelope{
				Variant: "info",
			},
			want: false,
		},
		{
			name: "payload field access",
			expr: `payload.severity == "HIGH"`,
			envelope: types.MappingEnvelope{
				Variant: "alert",
				Payload: json.RawMessage(`{"severity":"HIGH"}`),
			},
			want: true,
		},
		{
			name: "payload field access non-matching",
			expr: `payload.severity == "HIGH"`,
			envelope: types.MappingEnvelope{
				Payload: json.RawMessage(`{"severity":"LOW"}`),
			},
			want: false,
		},
		{
			name:    "invalid expression returns ErrFilterExprEval",
			expr:    `???invalid`,
			wantErr: ErrFilterExprEval,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := EvalFilter(ctx, tc.expr, tc.envelope)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}

				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got != tc.want {
				t.Fatalf("EvalFilter() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEvalMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		expr     string
		envelope types.MappingEnvelope
		wantKey  string
		wantVal  string
		wantErr  error
		wantRaw  bool
	}{
		{
			name: "empty expression returns original payload",
			expr: "",
			envelope: types.MappingEnvelope{
				Payload: json.RawMessage(`{"original":true}`),
			},
			wantRaw: true,
		},
		{
			name: "transform expression produces mapped output",
			expr: `{"mapped_variant": variant}`,
			envelope: types.MappingEnvelope{
				Variant: "alert",
				Payload: json.RawMessage(`{}`),
			},
			wantKey: "mapped_variant",
			wantVal: "alert",
		},
		{
			name:    "invalid expression returns ErrMapExprEval",
			expr:    `???invalid`,
			wantErr: ErrMapExprEval,
		},
	}

	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := EvalMap(ctx, tc.expr, tc.envelope)
			if tc.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error wrapping %v, got nil", tc.wantErr)
				}

				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tc.wantErr, err)
				}

				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantRaw {
				if string(got) != string(tc.envelope.Payload) {
					t.Fatalf("expected original payload %s, got %s", tc.envelope.Payload, got)
				}

				return
			}

			var m map[string]any
			if err := json.Unmarshal(got, &m); err != nil {
				t.Fatalf("failed to unmarshal result: %v", err)
			}

			val, ok := m[tc.wantKey]
			if !ok {
				t.Fatalf("expected key %q in result %s", tc.wantKey, got)
			}

			if val != tc.wantVal {
				t.Fatalf("expected %q=%q, got %q", tc.wantKey, tc.wantVal, val)
			}
		})
	}
}

func TestEvalFilter_NonBoolResult(t *testing.T) {
	t.Parallel()

	// Expression that returns a string instead of a bool should return ErrFilterExprEval
	ctx := context.Background()
	_, err := EvalFilter(ctx, `"not-a-bool"`, types.MappingEnvelope{})

	if err == nil {
		t.Fatal("expected error for non-bool result")
	}

	if !errors.Is(err, ErrFilterExprEval) {
		t.Fatalf("expected ErrFilterExprEval, got %v", err)
	}
}

func TestEvalMap_PayloadTransform(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	env := types.MappingEnvelope{
		Variant:  "alert",
		Resource: "vulns",
		Action:   "created",
		Payload:  json.RawMessage(`{"severity":"HIGH","id":42}`),
	}

	got, err := EvalMap(ctx, `{"sev": payload.severity, "src": resource}`, env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(got, &m); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	if m["sev"] != "HIGH" {
		t.Fatalf("expected sev=HIGH, got %v", m["sev"])
	}

	if m["src"] != "vulns" {
		t.Fatalf("expected src=vulns, got %v", m["src"])
	}
}

func TestEnvelopeToVars(t *testing.T) {
	t.Parallel()

	t.Run("all fields populated", func(t *testing.T) {
		t.Parallel()

		env := types.MappingEnvelope{
			Variant:  "alert",
			Resource: "vulnerabilities",
			Action:   "created",
			Payload:  json.RawMessage(`{"id":42,"name":"test"}`),
		}

		vars := envelopeToVars(env)

		if vars[celVarVariant] != "alert" {
			t.Fatalf("variant = %v, want %q", vars[celVarVariant], "alert")
		}

		if vars[celVarResource] != "vulnerabilities" {
			t.Fatalf("resource = %v, want %q", vars[celVarResource], "vulnerabilities")
		}

		if vars[celVarAction] != "created" {
			t.Fatalf("action = %v, want %q", vars[celVarAction], "created")
		}

		payload, ok := vars[celVarPayload]
		if !ok {
			t.Fatal("payload key missing from vars")
		}

		payloadMap, ok := payload.(map[string]any)
		if !ok {
			t.Fatalf("payload expected map[string]any, got %T", payload)
		}

		if payloadMap["name"] != "test" {
			t.Fatalf("payload.name = %v, want %q", payloadMap["name"], "test")
		}

		envelopeVar, ok := vars[celVarEnvelope].(map[string]any)
		if !ok {
			t.Fatal("envelope var should be map[string]any")
		}

		if envelopeVar["variant"] != "alert" {
			t.Fatalf("envelope.variant = %v, want %q", envelopeVar["variant"], "alert")
		}
	})

	t.Run("invalid JSON payload falls back to string", func(t *testing.T) {
		t.Parallel()

		env := types.MappingEnvelope{
			Variant: "raw",
			Payload: json.RawMessage(`not-valid-json`),
		}

		vars := envelopeToVars(env)

		payload, ok := vars[celVarPayload].(string)
		if !ok {
			t.Fatalf("expected string payload fallback, got %T", vars[celVarPayload])
		}

		if payload != "not-valid-json" {
			t.Fatalf("payload = %q, want %q", payload, "not-valid-json")
		}
	})

	t.Run("empty payload", func(t *testing.T) {
		t.Parallel()

		env := types.MappingEnvelope{
			Variant: "info",
		}

		vars := envelopeToVars(env)

		if vars[celVarPayload] != nil {
			t.Fatalf("expected nil payload, got %v", vars[celVarPayload])
		}

		if vars[celVarVariant] != "info" {
			t.Fatalf("variant = %v, want %q", vars[celVarVariant], "info")
		}
	})
}
