package slack

import (
	"golang.org/x/oauth2"

	"github.com/theopenlane/core/common/models"
	"github.com/theopenlane/core/pkg/jsonx"
)

// slackCredentialProviderData captures provider metadata extracted from Slack OAuth tokens
type slackCredentialProviderData struct {
	// AppID is the Slack app identifier
	AppID string `json:"appId,omitempty"`
	// BotUserID is the bot user identifier
	BotUserID string `json:"botUserId,omitempty"`
	// TeamID is the workspace identifier
	TeamID string `json:"teamId,omitempty"`
	// TeamName is the workspace display name
	TeamName string `json:"teamName,omitempty"`
}

// slackTokenTeam captures the nested Slack `team` token field
type slackTokenTeam struct {
	// ID is the workspace identifier
	ID string `json:"id"`
	// Name is the workspace display name
	Name string `json:"name"`
}

// slackCredentialSet extracts workspace metadata from Slack OAuth token response fields
func slackCredentialSet(token *oauth2.Token) models.CredentialSet {
	if token == nil {
		return models.CredentialSet{}
	}

	providerData := slackCredentialProviderData{}

	if appID, ok := token.Extra("app_id").(string); ok && appID != "" {
		providerData.AppID = appID
	}

	if botUserID, ok := token.Extra("bot_user_id").(string); ok && botUserID != "" {
		providerData.BotUserID = botUserID
	}

	if teamID, ok := token.Extra("team_id").(string); ok && teamID != "" {
		providerData.TeamID = teamID
	}

	if teamName, ok := token.Extra("team_name").(string); ok && teamName != "" {
		providerData.TeamName = teamName
	}

	teamRaw := token.Extra("team")
	if teamRaw != nil {
		var team slackTokenTeam
		if err := jsonx.RoundTrip(teamRaw, &team); err == nil {
			if team.ID != "" {
				providerData.TeamID = team.ID
			}
			if team.Name != "" {
				providerData.TeamName = team.Name
			}
		}
	}

	providerDataMap, err := jsonx.ToMap(providerData)
	if err != nil || len(providerDataMap) == 0 {
		return models.CredentialSet{}
	}

	return models.CredentialSet{
		ProviderData: providerDataMap,
	}
}
