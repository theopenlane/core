package awssecurityhub

import (
	"context"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamtypes "github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/theopenlane/core/internal/ent/integrationgenerated"
	"github.com/theopenlane/core/internal/integrations/providerkit"
	"github.com/theopenlane/core/internal/integrations/types"
	"github.com/theopenlane/core/pkg/logx"
)

const iamPageSize = int32(100)

// iamUserPayload is the JSON-serializable representation of one IAM user
type iamUserPayload struct {
	// ID is the stable IAM user identifier
	ID string `json:"id"`
	// Arn is the Amazon Resource Name for the user
	Arn string `json:"arn"`
	// UserName is the friendly IAM username
	UserName string `json:"userName"`
	// Path is the IAM path prefix for the user
	Path string `json:"path"`
	// Tags is a map of IAM tag key-value pairs attached to the user
	Tags map[string]string `json:"tags,omitempty"`
}

// iamGroupPayload is the JSON-serializable representation of one IAM group
type iamGroupPayload struct {
	// ID is the stable IAM group identifier
	ID string `json:"id"`
	// Arn is the Amazon Resource Name for the group
	Arn string `json:"arn"`
	// Name is the friendly IAM group name
	Name string `json:"name"`
	// Path is the IAM path prefix for the group
	Path string `json:"path"`
}

// iamEntityRef is a lightweight reference to an IAM user or group
type iamEntityRef struct {
	// ID is the stable IAM entity identifier
	ID string `json:"id"`
	// Name is the IAM entity name
	Name string `json:"name"`
}

// iamMembershipPayload is the envelope payload for one IAM group membership record
type iamMembershipPayload struct {
	// Group is the group side of the membership
	Group iamEntityRef `json:"group"`
	// Member is the user side of the membership
	Member iamEntityRef `json:"member"`
}

// IngestHandle adapts IAM directory sync to the ingest operation registration boundary
func (d DirectorySync) IngestHandle() types.IngestHandler {
	return providerkit.WithClientRequestConfig(iamClient, directorySyncOperation, ErrOperationConfigInvalid, func(ctx context.Context, _ types.OperationRequest, client *iam.Client, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
		if cfg.Disable {
			logx.FromContext(ctx).Debug().Msg("aws_iam: directory sync is disabled")

			return nil, nil
		}

		return d.Run(ctx, client, cfg)
	})
}

// Run collects AWS IAM users, and optionally groups and memberships
func (DirectorySync) Run(ctx context.Context, client *iam.Client, cfg DirectorySync) ([]types.IngestPayloadSet, error) {
	users, err := listIAMUsers(ctx, client)
	if err != nil {
		return nil, err
	}

	accountEnvelopes := make([]types.MappingEnvelope, 0, len(users))

	for _, user := range users {
		payload := iamUserToPayload(user)

		envelope, err := providerkit.MarshalEnvelope(payload.ID, payload, ErrDirectorySyncPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("user", payload.UserName).Msg("awsiam: failed to marshal user")
			return nil, err
		}

		accountEnvelopes = append(accountEnvelopes, envelope)
	}

	payloadSets := []types.IngestPayloadSet{
		{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryAccount,
			Envelopes: accountEnvelopes,
		},
	}

	if !cfg.DisableGroupSync {
		logx.FromContext(ctx).Info().Int("user_count", len(accountEnvelopes)).Msg("awsiam: collected IAM users")
		return payloadSets, nil
	}

	groups, err := listIAMGroups(ctx, client)
	if err != nil {
		return nil, err
	}

	groupEnvelopes := make([]types.MappingEnvelope, 0, len(groups))

	for _, group := range groups {
		payload := iamGroupToPayload(group)

		envelope, err := providerkit.MarshalEnvelope(payload.ID, payload, ErrDirectorySyncPayloadEncode)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("group", payload.Name).Msg("awsiam: failed to marshal group")
			return nil, err
		}

		groupEnvelopes = append(groupEnvelopes, envelope)
	}

	membershipEnvelopes := make([]types.MappingEnvelope, 0)

	for _, user := range users {
		userPayload := iamUserToPayload(user)
		memberRef := iamEntityRef{ID: userPayload.ID, Name: userPayload.UserName}

		userGroups, err := listGroupsForUser(ctx, client, userPayload.UserName)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("user", userPayload.UserName).Msg("awsiam: failed to list groups for user")
			return nil, err
		}

		for _, g := range userGroups {
			groupRef := iamEntityRef{
				ID:   awssdk.ToString(g.GroupId),
				Name: awssdk.ToString(g.GroupName),
			}

			envelope, err := providerkit.MarshalEnvelope(groupRef.ID+":"+memberRef.ID, iamMembershipPayload{
				Group:  groupRef,
				Member: memberRef,
			}, ErrDirectorySyncPayloadEncode)
			if err != nil {
				return nil, err
			}

			membershipEnvelopes = append(membershipEnvelopes, envelope)
		}
	}

	logx.FromContext(ctx).Debug().
		Int("user_count", len(accountEnvelopes)).
		Int("group_count", len(groupEnvelopes)).
		Int("membership_count", len(membershipEnvelopes)).
		Msg("awsiam: collected IAM users, groups, and memberships")

	payloadSets = append(payloadSets,
		types.IngestPayloadSet{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryGroup,
			Envelopes: groupEnvelopes,
		},
		types.IngestPayloadSet{
			Schema:    integrationgenerated.IntegrationMappingSchemaDirectoryMembership,
			Envelopes: membershipEnvelopes,
		},
	)

	return payloadSets, nil
}

// iamUserToPayload maps an IAM User SDK type to a JSON-serializable payload struct
func iamUserToPayload(user iamtypes.User) iamUserPayload {
	payload := iamUserPayload{
		ID:       awssdk.ToString(user.UserId),
		Arn:      awssdk.ToString(user.Arn),
		UserName: awssdk.ToString(user.UserName),
		Path:     awssdk.ToString(user.Path),
	}

	if len(user.Tags) > 0 {
		tags := make(map[string]string, len(user.Tags))
		for _, t := range user.Tags {
			if t.Key != nil && t.Value != nil {
				tags[*t.Key] = *t.Value
			}
		}
		payload.Tags = tags
	}

	return payload
}

// iamGroupToPayload maps an IAM Group SDK type to a JSON-serializable payload struct
func iamGroupToPayload(group iamtypes.Group) iamGroupPayload {
	return iamGroupPayload{
		ID:   awssdk.ToString(group.GroupId),
		Arn:  awssdk.ToString(group.Arn),
		Name: awssdk.ToString(group.GroupName),
		Path: awssdk.ToString(group.Path),
	}
}

// listIAMUsers pages through all IAM users using Marker-based pagination and
// fetches tags for each user separately
func listIAMUsers(ctx context.Context, client *iam.Client) ([]iamtypes.User, error) {
	var users []iamtypes.User
	input := &iam.ListUsersInput{MaxItems: awssdk.Int32(iamPageSize)}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := client.ListUsers(ctx, input)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("awsiam: error listing IAM users")
			return nil, ErrIAMUsersFetchFailed
		}

		users = append(users, resp.Users...)

		if !resp.IsTruncated {
			break
		}

		input.Marker = resp.Marker
	}

	for i, user := range users {
		tags, err := listUserTags(ctx, client, awssdk.ToString(user.UserName))
		if err != nil {
			return nil, err
		}

		users[i].Tags = tags
	}

	return users, nil
}

// listUserTags pages through all tags for a single IAM user
func listUserTags(ctx context.Context, client *iam.Client, userName string) ([]iamtypes.Tag, error) {
	var tags []iamtypes.Tag
	input := &iam.ListUserTagsInput{UserName: awssdk.String(userName)}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := client.ListUserTags(ctx, input)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("userName", userName).Msg("awsiam: error listing tags for user")
			return nil, ErrIAMUsersFetchFailed
		}

		tags = append(tags, resp.Tags...)

		if !resp.IsTruncated {
			break
		}

		input.Marker = resp.Marker
	}

	return tags, nil
}

// listIAMGroups pages through all IAM groups using Marker-based pagination
func listIAMGroups(ctx context.Context, client *iam.Client) ([]iamtypes.Group, error) {
	var groups []iamtypes.Group
	input := &iam.ListGroupsInput{MaxItems: awssdk.Int32(iamPageSize)}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := client.ListGroups(ctx, input)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Msg("awsiam: error listing IAM groups")
			return nil, ErrIAMGroupsFetchFailed
		}

		groups = append(groups, resp.Groups...)

		if !resp.IsTruncated {
			break
		}

		input.Marker = resp.Marker
	}

	return groups, nil
}

// listGroupsForUser pages through all IAM groups for one user using Marker-based pagination
func listGroupsForUser(ctx context.Context, client *iam.Client, userName string) ([]iamtypes.Group, error) {
	var groups []iamtypes.Group
	input := &iam.ListGroupsForUserInput{
		UserName: awssdk.String(userName),
		MaxItems: awssdk.Int32(iamPageSize),
	}

	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		resp, err := client.ListGroupsForUser(ctx, input)
		if err != nil {
			logx.FromContext(ctx).Error().Err(err).Str("userName", userName).Msg("awsiam: error listing groups for user")
			return nil, ErrIAMGroupsForUserFetchFailed
		}

		groups = append(groups, resp.Groups...)

		if !resp.IsTruncated {
			break
		}

		input.Marker = resp.Marker
	}

	return groups, nil
}
