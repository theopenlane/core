package gencmd

import (
	"reflect"
	"testing"
)

func TestExtractNodeFieldPaths(t *testing.T) {
	graphql := `query Integrations {
	  integrations {
	    edges {
	      node {
	        id
	        name
	        owner {
	          id
	          displayName
	        }
	      }
	    }
	  }
	}`

	paths := extractNodeFieldPaths(graphql, "integrations")
	expected := [][]string{
		{"id"},
		{"name"},
		{"owner", "id"},
		{"owner", "displayName"},
	}

	if !reflect.DeepEqual(paths, expected) {
		t.Fatalf("unexpected paths: %#v", paths)
	}
}
