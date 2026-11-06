package placeholder

import (
	"encoding/json"
	"testing"

	"gotest.tools/v3/assert"
)

func TestVariablesFill(t *testing.T) {
	variables := Variables{"STATUS": "DRAFT", "NAME": `Acme "Quoted"`, "EMPTY": "", "object.ref_code": "CC1.1"}

	testCases := []struct {
		name     string
		doc      string
		expected string
	}{
		{name: "set variable replaces the default", doc: `{"status":"{{STATUS|PUBLISHED}}"}`, expected: `{"status":"DRAFT"}`},
		{name: "unset variable keeps the default", doc: `{"approver":"{{APPROVER|Admins}}"}`, expected: `{"approver":"Admins"}`},
		{name: "empty variable empties the value", doc: `{"approver":"{{EMPTY|Admins}}"}`, expected: `{"approver":""}`},
		{name: "empty default", doc: `{"note":"{{NOTE|}}"}`, expected: `{"note":""}`},
		{name: "value is escaped", doc: `{"name":"{{NAME|x}}"}`, expected: `{"name":"Acme \"Quoted\""}`},
		{name: "variable without a default", doc: `{"status":"{{STATUS}}"}`, expected: `{"status":"DRAFT"}`},
		{name: "unset variable without a default is untouched", doc: `{"name":"{{ORG_NAME}}"}`, expected: `{"name":"{{ORG_NAME}}"}`},
		{name: "lowercase dotted name", doc: `{"ref":"{{object.ref_code}}"}`, expected: `{"ref":"CC1.1"}`},
		{name: "unset variable with an offset is untouched", doc: `{"due":"{{TODAY+1}}"}`, expected: `{"due":"{{TODAY+1}}"}`},
		{name: "offset on a value that is not a date is untouched", doc: `{"status":"{{STATUS+1}}"}`, expected: `{"status":"{{STATUS+1}}"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, variables.Fill(tc.doc))
		})
	}
}

func TestVariablesWith(t *testing.T) {
	base := Variables{"ORG_NAME": "Acme", "ORG_DOMAIN": "acme.test"}

	merged := base.With(Variables{"ORG_DOMAIN": "override.test", "STATUS": "DRAFT"})

	assert.DeepEqual(t, Variables{"ORG_NAME": "Acme", "ORG_DOMAIN": "override.test", "STATUS": "DRAFT"}, merged)
	assert.Equal(t, "acme.test", base["ORG_DOMAIN"])
	assert.DeepEqual(t, Variables{"STATUS": "DRAFT"}, Variables(nil).With(Variables{"STATUS": "DRAFT"}))
}

func TestFill(t *testing.T) {
	variables := Variables{"ORG_NAME": `Acme "Quoted" \ Co`, "ORG_DOMAIN": "acme.test", "TODAY": "2026-11-11"}

	testCases := []struct {
		name     string
		doc      string
		expected string
	}{
		{name: "org name is escaped", doc: `{"name":"{{ORG_NAME}} Security Policy"}`, expected: `{"name":"Acme \"Quoted\" \\ Co Security Policy"}`},
		{name: "org domain", doc: `{"email":"security@{{ORG_DOMAIN}}"}`, expected: `{"email":"security@acme.test"}`},
		{name: "today", doc: `{"due":"{{TODAY}}"}`, expected: `{"due":"2026-11-11"}`},
		{name: "days ahead", doc: `{"due":"{{TODAY+14}}"}`, expected: `{"due":"2026-11-25"}`},
		{name: "days behind", doc: `{"due":"{{TODAY-41}}"}`, expected: `{"due":"2026-10-01"}`},
		{name: "keeps time of day", doc: `{"reviewed_at":"{{TODAY-1}}T09:15:00Z"}`, expected: `{"reviewed_at":"2026-11-10T09:15:00Z"}`},
		{name: "several in one document", doc: `{"start":"{{TODAY-2}}","end":"{{TODAY+2}}"}`, expected: `{"start":"2026-11-09","end":"2026-11-13"}`},
		{name: "variable and date together", doc: `{"name":"{{ORG_NAME}} review","due":"{{TODAY+1}}"}`, expected: `{"name":"Acme \"Quoted\" \\ Co review","due":"2026-11-12"}`},
		{name: "literal dates stay fixed", doc: `{"effective":"2026-01-23","details":"as of {{TODAY}}"}`, expected: `{"effective":"2026-01-23","details":"as of 2026-11-11"}`},
		{name: "offset too large stays a placeholder", doc: `{"due":"{{TODAY+99999999999999999999}}"}`, expected: `{"due":"{{TODAY+99999999999999999999}}"}`},
		{name: "malformed placeholder untouched", doc: `{"due":"{{TODAY+}}"}`, expected: `{"due":"{{TODAY+}}"}`},
		{name: "unset variable keeps its default", doc: `{"owner":"{{ORG_OWNER_NAME|Someone}}"}`, expected: `{"owner":"Someone"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filled := variables.Fill(tc.doc)

			assert.Equal(t, tc.expected, filled)
			assert.Check(t, json.Valid([]byte(filled)))
		})
	}
}
