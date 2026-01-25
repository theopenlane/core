package graphapi

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"

	"github.com/samber/lo"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	gentemplate "github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/graphapi/common"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
)

//go:embed trustcenternda.json.tpl
var trustCenterNDATemplate string

var errOneNDAOnly = errors.New("one NDA file is required")

func createTrustCenterNDA(ctx context.Context, input model.CreateTrustCenterNDAInput) (*model.TrustCenterNDACreatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	trustCenter, err := txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	// set the organization in the auth context if its not done for us
	if err := common.SetOrganizationInAuthContext(ctx, &trustCenter.OwnerID); err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to set organization in auth context")

		return nil, rout.ErrPermissionDenied
	}

	templateObj, err := txnCtx.Template.Create().
		SetInput(
			generated.CreateTemplateInput{
				Name:          "Trust Center NDA",
				TemplateType:  &enums.Document,
				Kind:          &enums.TemplateKindTrustCenterNda,
				Jsonconfig:    map[string]interface{}{},
				OwnerID:       &trustCenter.OwnerID,
				TrustCenterID: &trustCenter.ID,
			},
		).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "template"})
	}

	// Parse the template
	tmpl, err := template.New("nda").Parse(trustCenterNDATemplate)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	key := "templateFiles"

	// get the file from the context, if it exists
	files, _ := objects.FilesFromContextWithKey(ctx, key)

	if len(files) != 1 {
		return nil, parseRequestError(ctx, errOneNDAOnly, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	// Define the data to be used in the template
	data := struct {
		TrustCenterID string
		NDAFileID     string
	}{
		TrustCenterID: trustCenter.ID,
		NDAFileID:     files[0].ID,
	}

	// Create a bytes.Buffer to capture the output
	var buf bytes.Buffer

	// Execute the template, writing the output to the buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	// Get the output as a string from the buffer
	outputString := buf.String()

	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(templateObj).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	return &model.TrustCenterNDACreatePayload{
		Template: updatedTmpl,
	}, nil
}

func updateTrustCenterNDA(ctx context.Context, id string) (*model.TrustCenterNDAUpdatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	ndaTemplate, err := txnCtx.Template.Query().Where(gentemplate.TrustCenterIDEQ(id)).Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	key := "templateFiles"
	// get the file from the context, if it exists
	files, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(files) != 1 {
		return &model.TrustCenterNDAUpdatePayload{
			Template: ndaTemplate,
		}, nil
	}

	// Parse the template
	tmpl, err := template.New("nda").Parse(trustCenterNDATemplate)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	// Define the data to be used in the template
	data := struct {
		TrustCenterID string
		NDAFileID     string
	}{
		TrustCenterID: ndaTemplate.TrustCenterID,
		NDAFileID:     files[0].ID,
	}

	// Create a bytes.Buffer to capture the output
	var buf bytes.Buffer

	// Execute the template, writing the output to the buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		logx.FromContext(ctx).Error().Err(err).Msg("failed to execute nda template")

		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	// Get the output as a string from the buffer
	outputString := buf.String()

	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(ndaTemplate).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	return &model.TrustCenterNDAUpdatePayload{
		Template: updatedTmpl,
	}, nil
}

// submitTrustCenterNDAResponse submits a trust center NDA response
func submitTrustCenterNDAResponse(ctx context.Context, input model.SubmitTrustCenterNDAResponseInput) (*model.SubmitTrustCenterNDAResponsePayload, error) {
	anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx)
	if !ok || anon.SubjectEmail == "" || anon.TrustCenterID == "" || anon.OrganizationID == "" {
		return nil, newPermissionDeniedError()
	}

	allowCtx := contextx.With(privacy.DecisionContext(ctx, privacy.Allow), auth.TrustCenterNDAContextKey{
		OrgID: anon.OrganizationID,
	})

	res, err := withTransactionalMutation(allowCtx).DocumentData.Create().SetInput(
		generated.CreateDocumentDataInput{
			TemplateID: lo.ToPtr(input.TemplateID),
			Data:       input.Response,
		},
	).Save(allowCtx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	return &model.SubmitTrustCenterNDAResponsePayload{
		DocumentData: res,
	}, nil
}
