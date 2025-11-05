package speccli

import (
	"encoding/json"
	"reflect"
	"testing"
	"testing/fstest"
)

type createInput struct {
	Name string
}

type updateInput struct {
	Name   *string
	Tags   []string
	Active *bool
}

type widgetWhereInput struct {
	Name *string
}

func TestLoadSpecsFromFS(t *testing.T) {
	spec := map[string]any{
		"name":  "widget",
		"use":   "widget",
		"short": "manage widgets",
		"list": map[string]any{
			"method":       "GetAllWidgets",
			"root":         "widgets",
			"filterMethod": "GetWidgets",
			"where": map[string]any{
				"type": "example.WidgetWhereInput",
				"fields": []any{
					map[string]any{
						"flag":  map[string]any{"name": "name", "usage": "widget name"},
						"field": "Name",
					},
				},
			},
		},
		"get": map[string]any{
			"method":       "GetWidgetByID",
			"idFlag":       map[string]any{"name": "id", "usage": "widget id"},
			"resultPath":   []string{"widget"},
			"fallbackList": true,
		},
		"create": map[string]any{
			"method":     "CreateWidget",
			"inputType":  "example.CreateWidgetInput",
			"resultPath": []string{"createWidget", "widget"},
			"fields": []any{
				map[string]any{
					"flag": map[string]any{
						"name":     "name",
						"usage":    "widget name",
						"required": true,
					},
					"field": "Name",
					"kind":  "string",
				},
			},
		},
		"update": map[string]any{
			"method":     "UpdateWidget",
			"idFlag":     map[string]any{"name": "id", "usage": "widget id"},
			"inputType":  "example.UpdateWidgetInput",
			"resultPath": []string{"updateWidget", "widget"},
			"fields": []any{
				map[string]any{
					"flag":  map[string]any{"name": "name", "usage": "widget name"},
					"field": "Name",
					"kind":  "string",
				},
				map[string]any{
					"flag":  map[string]any{"name": "tags", "usage": "widget tags"},
					"field": "Tags",
					"kind":  "stringSlice",
				},
			},
		},
		"delete": map[string]any{
			"method":      "DeleteWidget",
			"idFlag":      map[string]any{"name": "id", "usage": "widget id"},
			"resultPath":  []string{"deleteWidget", "deletedID"},
			"resultField": "deletedID",
		},
		"columns": []any{
			map[string]any{"header": "ID", "path": []string{"id"}},
			map[string]any{"header": "Name", "path": []string{"name"}},
		},
		"deleteColumns": []any{
			map[string]any{"header": "DeletedID", "path": []string{"deletedID"}},
		},
	}

	payload, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}

	fsys := fstest.MapFS{
		"widget.json": {
			Data: payload,
		},
	}

	resolver := StaticTypeResolver(map[string]reflect.Type{
		"example.CreateWidgetInput": reflect.TypeOf(createInput{}),
		"example.UpdateWidgetInput": reflect.TypeOf(updateInput{}),
		"example.WidgetWhereInput":  reflect.TypeOf(widgetWhereInput{}),
	})

	specs, err := LoadSpecsFromFS(fsys, LoaderOptions{
		TypeResolver: resolver,
	})
	if err != nil {
		t.Fatalf("LoadSpecsFromFS: %v", err)
	}

	if len(specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(specs))
	}

	s := specs[0]
	if s.Name != "widget" || s.Use != "widget" || s.Short != "manage widgets" {
		t.Fatalf("unexpected metadata: %+v", s)
	}

	if s.Create == nil || s.Update == nil || s.Delete == nil || s.Get == nil || s.List == nil {
		t.Fatalf("expected CRUD spec, got %+v", s)
	}

	if s.Create.InputType != reflect.TypeOf(createInput{}) {
		t.Fatalf("create input type mismatch: %v", s.Create.InputType)
	}

	if s.Update.InputType != reflect.TypeOf(updateInput{}) {
		t.Fatalf("update input type mismatch: %v", s.Update.InputType)
	}

	if len(s.Columns) != 2 || s.Columns[0].Header != "ID" {
		t.Fatalf("unexpected columns: %+v", s.Columns)
	}

	if s.List.FilterMethod != "GetWidgets" {
		t.Fatalf("unexpected filter method: %s", s.List.FilterMethod)
	}

	if s.List.Where == nil || len(s.List.Where.Fields) != 1 {
		t.Fatalf("unexpected where spec: %+v", s.List.Where)
	}
}

func TestLoadSpecsFromFSOverride(t *testing.T) {
	spec := map[string]any{
		"name":  "widget",
		"use":   "widget",
		"short": "manage widgets",
		"create": map[string]any{
			"method":     "CreateWidget",
			"inputType":  "example.CreateWidgetInput",
			"resultPath": []string{"createWidget", "widget"},
			"fields": []any{
				map[string]any{
					"flag": map[string]any{
						"name":     "name",
						"usage":    "widget name",
						"required": true,
					},
					"field": "Name",
				},
			},
		},
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}

	fsys := fstest.MapFS{
		"widget.json": {
			Data: payload,
		},
	}

	overrideCalled := false

	_, err = LoadSpecsFromFS(fsys, LoaderOptions{
		TypeResolver: StaticTypeResolver(map[string]reflect.Type{
			"example.CreateWidgetInput": reflect.TypeOf(createInput{}),
		}),
		Overrides: map[string]SpecOverride{
			"widget": func(spec *CommandSpec) error {
				overrideCalled = true
				spec.Short = "overridden"
				return nil
			},
		},
	})
	if err != nil {
		t.Fatalf("LoadSpecsFromFS: %v", err)
	}

	if !overrideCalled {
		t.Fatalf("override was not invoked")
	}
}

func TestLoadSpecsFromFSMissingType(t *testing.T) {
	spec := map[string]any{
		"name":  "widget",
		"use":   "widget",
		"short": "manage widgets",
		"create": map[string]any{
			"method":     "CreateWidget",
			"inputType":  "example.CreateWidgetInput",
			"resultPath": []string{"createWidget", "widget"},
		},
	}

	payload, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}

	fsys := fstest.MapFS{
		"widget.json": {
			Data: payload,
		},
	}

	_, err = LoadSpecsFromFS(fsys, LoaderOptions{})
	if err == nil {
		t.Fatalf("expected error for missing type resolver")
	}
}
