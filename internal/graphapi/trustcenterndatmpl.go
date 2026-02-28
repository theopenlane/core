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
	"github.com/theopenlane/utils/rout"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	gentemplate "github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/generated/trustcenterndarequest"
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

	var (
		trustCenter *generated.TrustCenter
		err         error
	)

	if input.TrustCenterID == "" {
		// get the trust center
		trustCenter, err = txnCtx.TrustCenter.Query().Only(ctx)
		if err != nil {
			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
		}
	} else {
		trustCenter, err = txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
		if err != nil {
			return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
		}
	}

	// set the organization in the auth context if its not done for us
	ctx, err = common.SetOrganizationInAuthContext(ctx, &trustCenter.OwnerID)
	if err != nil {
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

	ndaTemplate, err := txnCtx.Template.Query().
		Where(gentemplate.TrustCenterIDEQ(id)).
		Where(gentemplate.KindEQ(enums.TemplateKindTrustCenterNda)).
		Only(ctx)
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

	updatedTmpl, err := txnCtx.Template.UpdateOne(ndaTemplate).
		SetTrustCenterID(id). // needed so the hook can access this
		SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionUpdate, Object: "trustcenternda"})
	}

	return &model.TrustCenterNDAUpdatePayload{
		Template: updatedTmpl,
	}, nil
}

// submitTrustCenterNDAResponse submits a trust center NDA response
func submitTrustCenterNDAResponse(ctx context.Context, input model.SubmitTrustCenterNDAResponseInput) (*model.SubmitTrustCenterNDAResponsePayload, error) {
	tcID, hasTCID := auth.ActiveTrustCenterIDKey.Get(ctx)
	caller, hasCaller := auth.CallerFromContext(ctx)
	if !hasTCID || tcID == "" || !hasCaller || caller == nil || caller.SubjectEmail == "" || caller.OrganizationID == "" {
		return nil, newPermissionDeniedError()
	}

	allowCtx := auth.WithCaller(privacy.DecisionContext(ctx, privacy.Allow), caller)

	txnCtx := withTransactionalMutation(allowCtx)

	ndaRequest, err := txnCtx.TrustCenterNDARequest.Query().
		Where(
			trustcenterndarequest.EmailEqualFold(caller.SubjectEmail),
			trustcenterndarequest.TrustCenterID(tcID),
		).First(allowCtx)
	if err != nil {
		return nil, parseRequestError(ctx, err, common.Action{Action: common.ActionCreate, Object: "trustcenternda"})
	}

	// we need to update the signatory info from the NDA request we collected previously
	if signatoryInfo, ok := input.Response["signatory_info"].(map[string]any); ok {
		signatoryInfo["first_name"] = ndaRequest.FirstName
		signatoryInfo["last_name"] = ndaRequest.LastName
		signatoryInfo["company_name"] = lo.FromPtr(ndaRequest.CompanyName)
	}

	res, err := txnCtx.DocumentData.Create().SetInput(
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
