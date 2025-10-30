package scim

import (
	"fmt"
	"net/mail"
	"strings"

	"github.com/elimity-com/scim"
)

// UserAttributes holds extracted and validated SCIM user attributes
type UserAttributes struct {
	UserName          string
	Email             string
	ExternalID        string
	FirstName         string
	LastName          string
	DisplayName       string
	PreferredLanguage string
	Locale            string
	ProfileURL        string
	Active            bool
}

// ExtractUserAttributes extracts and validates user attributes from SCIM ResourceAttributes
func ExtractUserAttributes(attributes scim.ResourceAttributes) (*UserAttributes, error) {
	email, err := extractEmail(attributes)
	if err != nil {
		return nil, err
	}

	userName, _ := attributes["userName"].(string)
	if userName == "" {
		userName = email
	}

	ua := &UserAttributes{
		UserName: userName,
		Email:    email,
		Active:   true,
	}

	if externalID, ok := attributes["externalId"].(string); ok {
		ua.ExternalID = externalID
	}

	if preferredLanguage, ok := attributes["preferredLanguage"].(string); ok {
		ua.PreferredLanguage = preferredLanguage
	}

	if locale, ok := attributes["locale"].(string); ok {
		ua.Locale = locale
	}

	if profileURL, ok := attributes["profileUrl"].(string); ok {
		ua.ProfileURL = profileURL
	}

	if nameMap, ok := attributes["name"].(map[string]any); ok {
		if givenName, ok := nameMap["givenName"].(string); ok {
			ua.FirstName = givenName
		}
		if familyName, ok := nameMap["familyName"].(string); ok {
			ua.LastName = familyName
		}
	}

	if displayName, ok := attributes["displayName"].(string); ok {
		ua.DisplayName = displayName
	}

	if ua.DisplayName == "" {
		ua.DisplayName = strings.TrimSpace(fmt.Sprintf("%s %s", ua.FirstName, ua.LastName))
	}

	if ua.DisplayName == "" {
		ua.DisplayName = strings.ToLower(ua.Email)
	}

	if active, ok := attributes["active"].(bool); ok {
		ua.Active = active
	}

	return ua, nil
}

// PatchUserAttributes applies patch operations to update user attributes
type PatchUserAttributes struct {
	Email             *string
	UserName          *string
	ExternalID        *string
	PreferredLanguage *string
	Locale            *string
	ProfileURL        *string
	FirstName         *string
	LastName          *string
	DisplayName       *string
	Active            *bool
}

// ExtractPatchUserAttribute extracts a single attribute from a patch operation value
func ExtractPatchUserAttribute(op scim.PatchOperation) (*PatchUserAttributes, error) {
	patch := &PatchUserAttributes{}

	valueMap, isMap := op.Value.(map[string]any)
	if !isMap {
		return patch, nil
	}

	if externalID, ok := valueMap["externalId"].(string); ok {
		patch.ExternalID = &externalID
	}

	if preferredLanguage, ok := valueMap["preferredLanguage"].(string); ok {
		patch.PreferredLanguage = &preferredLanguage
	}

	if locale, ok := valueMap["locale"].(string); ok {
		patch.Locale = &locale
	}

	if profileURL, ok := valueMap["profileUrl"].(string); ok {
		patch.ProfileURL = &profileURL
	}

	if userName, ok := valueMap["userName"].(string); ok && userName != "" {
		if _, err := mail.ParseAddress(userName); err != nil {
			return nil, fmt.Errorf("%w: userName must be a valid email address", ErrInvalidAttributes)
		}
		patch.UserName = &userName
		patch.Email = &userName
	}

	if emailsArray, ok := valueMap["emails"].([]any); ok && len(emailsArray) > 0 {
		for _, emailItem := range emailsArray {
			emailMap, ok := emailItem.(map[string]any)
			if !ok {
				continue
			}

			value, ok := emailMap["value"].(string)
			if !ok || value == "" {
				continue
			}

			if _, err := mail.ParseAddress(value); err != nil {
				continue
			}

			if primary, ok := emailMap["primary"].(bool); ok && primary {
				patch.Email = &value
				break
			}
		}
	}

	if displayName, ok := valueMap["displayName"].(string); ok {
		patch.DisplayName = &displayName
	}

	if name, ok := valueMap["name"].(map[string]any); ok {
		if givenName, ok := name["givenName"].(string); ok {
			patch.FirstName = &givenName
		}

		if familyName, ok := name["familyName"].(string); ok {
			patch.LastName = &familyName
		}
	}

	if active, ok := valueMap["active"].(bool); ok {
		patch.Active = &active
	}

	return patch, nil
}

// GroupAttributes holds extracted and validated SCIM group attributes
type GroupAttributes struct {
	DisplayName string
	ExternalID  string
	MemberIDs   []string
	Active      bool
}

// ExtractGroupAttributes extracts and validates group attributes from SCIM ResourceAttributes
func ExtractGroupAttributes(attributes scim.ResourceAttributes) (*GroupAttributes, error) {
	displayName, ok := attributes["displayName"].(string)
	if !ok || displayName == "" {
		return nil, fmt.Errorf("%w: displayName is required", ErrInvalidAttributes)
	}

	ga := &GroupAttributes{
		DisplayName: displayName,
		Active:      true,
	}

	if externalID, ok := attributes["externalId"].(string); ok {
		ga.ExternalID = externalID
	}

	if active, ok := attributes["active"].(bool); ok {
		ga.Active = active
	}

	ga.MemberIDs = extractMemberIDsFromValue(attributes["members"])

	return ga, nil
}

// PatchGroupAttributes holds attributes that can be patched on a group
type PatchGroupAttributes struct {
	DisplayName *string
	ExternalID  *string
	Active      *bool
}

// ExtractPatchGroupAttribute extracts group attributes from a patch operation value
func ExtractPatchGroupAttribute(op scim.PatchOperation) *PatchGroupAttributes {
	patch := &PatchGroupAttributes{}

	valueMap, isMap := op.Value.(map[string]any)
	if !isMap {
		return patch
	}

	if externalID, ok := valueMap["externalId"].(string); ok {
		patch.ExternalID = &externalID
	}

	if displayName, ok := valueMap["displayName"].(string); ok {
		patch.DisplayName = &displayName
	}

	if active, ok := valueMap["active"].(bool); ok {
		patch.Active = &active
	}

	return patch
}

// extractMemberIDsFromValue extracts and deduplicates member IDs from a SCIM members value
func extractMemberIDsFromValue(value any) []string {
	members, ok := value.([]any)
	if !ok {
		return nil
	}

	memberIDs := make([]string, 0, len(members))

	for _, m := range members {
		memberMap, ok := m.(map[string]any)
		if !ok {
			continue
		}

		memberID, ok := memberMap["value"].(string)
		if !ok || memberID == "" {
			continue
		}

		memberIDs = append(memberIDs, memberID)
	}

	seen := make(map[string]bool)
	unique := make([]string, 0, len(memberIDs))

	for _, id := range memberIDs {
		if !seen[id] {
			seen[id] = true
			unique = append(unique, id)
		}
	}

	return unique
}
