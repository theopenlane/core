package entitlements

import (
	"reflect"
	"testing"
)

func TestNewQueryBuilder(t *testing.T) {
	qb := NewQueryBuilder()
	if qb == nil {
		t.Error("Expected NewQueryBuilder to return a non-nil QueryBuilder")
	}
	if len(qb.keys) != 0 {
		t.Errorf("Expected keys to be empty, got %v", qb.keys)
	}
}

func TestWithOperator(t *testing.T) {
	operator := " AND "
	qb := NewQueryBuilder(WithOperator(operator))
	if qb.operator != operator {
		t.Errorf("Expected operator to be %s, got %s", operator, qb.operator)
	}
}

func TestWithKeys(t *testing.T) {
	keys := map[string]string{"key1": "value1", "key2": "value2"}
	qb := NewQueryBuilder(WithKeys(keys))
	if !reflect.DeepEqual(qb.keys, keys) {
		t.Errorf("Expected keys to be %v, got %v", keys, qb.keys)
	}
}

func TestBuildQuery(t *testing.T) {
	keys := map[string]string{
		"key1":  "value1",
		"key2":  "value2",
		"key3":  "value3",
		"key4":  "value4",
		"key5":  "value5",
		"key6":  "value6",
		"key7":  "value7",
		"key8":  "value8",
		"key9":  "value9",
		"key10": "value10",
	}
	qb := NewQueryBuilder(WithOperator(" AND "), WithKeys(keys))
	expectedQueries := []string{
		"metadata['key1']:'value1' AND metadata['key2']:'value2' AND metadata['key3']:'value3' AND metadata['key4']:'value4' AND metadata['key5']:'value5' AND metadata['key6']:'value6' AND metadata['key7']:'value7' AND metadata['key8']:'value8' AND metadata['key9']:'value9' AND metadata['key10']:'value10'",
	}
	queries := qb.BuildQuery()
	if !reflect.DeepEqual(queries, expectedQueries) {
		t.Errorf("Expected queries to be %v, got %v", expectedQueries, queries)
	}
}
