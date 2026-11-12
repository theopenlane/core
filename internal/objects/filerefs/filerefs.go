package filerefs

import (
	"context"

	"github.com/samber/lo"

	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/actionplan"
	"github.com/theopenlane/core/v2/internal/ent/generated/directoryaccount"
	"github.com/theopenlane/core/v2/internal/ent/generated/entity"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/internalpolicy"
	"github.com/theopenlane/core/v2/internal/ent/generated/organization"
	"github.com/theopenlane/core/v2/internal/ent/generated/procedure"
	"github.com/theopenlane/core/v2/internal/ent/generated/standard"
	"github.com/theopenlane/core/v2/internal/ent/generated/subprocessor"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterentity"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterndarequest"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcentersetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/trustcenterwatermarkconfig"
	"github.com/theopenlane/core/v2/internal/ent/generated/user"
)

// lookup returns the ids referenced by one foreign key to files
type lookup func(ctx context.Context, client *ent.Client, ids []string) ([]string, error)

// lookups are the foreign keys to files with no edge on the file schema
var lookups = []lookup{
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.ActionPlan.Query().Where(actionplan.FileIDIn(ids...)).Select(actionplan.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.DirectoryAccount.Query().Where(directoryaccount.AvatarLocalFileIDIn(ids...)).Select(directoryaccount.FieldAvatarLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Entity.Query().Where(entity.LogoFileIDIn(ids...)).Select(entity.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.InternalPolicy.Query().Where(internalpolicy.FileIDIn(ids...)).Select(internalpolicy.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Organization.Query().Where(organization.AvatarLocalFileIDIn(ids...)).Select(organization.FieldAvatarLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Procedure.Query().Where(procedure.FileIDIn(ids...)).Select(procedure.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Standard.Query().Where(standard.LogoFileIDIn(ids...)).Select(standard.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.Subprocessor.Query().Where(subprocessor.LogoFileIDIn(ids...)).Select(subprocessor.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterEntity.Query().Where(trustcenterentity.LogoFileIDIn(ids...)).Select(trustcenterentity.FieldLogoFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterNDARequest.Query().Where(trustcenterndarequest.FileIDIn(ids...)).Select(trustcenterndarequest.FieldFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.FaviconLocalFileIDIn(ids...)).Select(trustcentersetting.FieldFaviconLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.HeroImageLocalFileIDIn(ids...)).Select(trustcentersetting.FieldHeroImageLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterSetting.Query().Where(trustcentersetting.LogoLocalFileIDIn(ids...)).Select(trustcentersetting.FieldLogoLocalFileID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.TrustCenterWatermarkConfig.Query().Where(trustcenterwatermarkconfig.LogoIDIn(ids...)).Select(trustcenterwatermarkconfig.FieldLogoID).Strings(ctx)
	},
	func(ctx context.Context, c *ent.Client, ids []string) ([]string, error) {
		return c.User.Query().Where(user.AvatarLocalFileIDIn(ids...)).Select(user.FieldAvatarLocalFileID).Strings(ctx)
	},
}

// Referenced returns the ids with an inbound reference; soft deleted parents count under entx.SkipSoftDelete
func Referenced(ctx context.Context, client *ent.Client, ids []string) (map[string]struct{}, error) {
	edged, err := client.File.Query().
		Where(file.IDIn(ids...), file.Or(
			file.HasGroups(),
			file.HasContact(),
			file.HasEntity(),
			file.HasOrganizationSetting(),
			file.HasTemplate(),
			file.HasDocument(),
			file.HasProgram(),
			file.HasPlatform(),
			file.HasEvidence(),
			file.HasIdentityHolder(),
			file.HasScan(),
			file.HasEvents(),
			file.HasIntegrations(),
			file.HasSecrets(),
			file.HasTrustCenterEntities(),
			file.HasTrustCenterDoc(),
			file.HasOriginalTrustCenterDoc(),
		)).
		IDs(ctx)
	if err != nil {
		return nil, err
	}

	referenced := lo.Keyify(edged)

	for _, lookup := range lookups {
		found, err := lookup(ctx, client, ids)
		if err != nil {
			return nil, err
		}

		for _, id := range found {
			referenced[id] = struct{}{}
		}
	}

	return referenced, nil
}
