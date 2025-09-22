package graphapi

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"html/template"

	"github.com/99designs/gqlgen/graphql"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	gentemplate "github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/utils/rout"
)

//go:embed trustcenternda.json.tpl
var trustCenterNDATemplate string

var (
	errOneNDAOnly          = errors.New("one NDA file is required")
	errNDATemplateNotFound = errors.New("NDA template not found")
	errNDATemplateRequired = errors.New("one NDA template is required")
)

func createTrustCenterNDA(ctx context.Context, input model.CreateTrustCenterNDAInput, _ []*graphql.Upload) (*model.TrustCenterNDACreatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	trustCenter, err := txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenternda"})
	}

	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, &trustCenter.OwnerID); err != nil {
		log.Error().Err(err).Msg("failed to set organization in auth context")

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
		return nil, parseRequestError(err, action{action: ActionCreate, object: "template"})
	}

	// Parse the template
	tmpl, err := template.New("nda").Parse(trustCenterNDATemplate)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenternda"})
	}
	key := "templateFiles"

	// get the file from the context, if it exists
	files, _ := objects.FilesFromContextWithKey(ctx, key)

	if len(files) != 1 {
		return nil, parseRequestError(errOneNDAOnly, action{action: ActionCreate, object: "trustcenternda"})
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
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenternda"})
	}

	// Get the output as a string from the buffer
	outputString := buf.String()
	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(templateObj).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenternda"})
	}

	return &model.TrustCenterNDACreatePayload{
		Template: updatedTmpl,
	}, nil

}

func updateTrustCenterNDA(ctx context.Context, id string, _ []*graphql.Upload) (*model.TrustCenterNDAUpdatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)
	templates, err := txnCtx.Template.Query().Where(gentemplate.TrustCenterIDEQ(id)).All(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	if len(templates) == 0 {
		return nil, parseRequestError(errNDATemplateNotFound, action{action: ActionUpdate, object: "trustcenternda"})
	}

	if len(templates) != 1 {
		return nil, parseRequestError(errNDATemplateRequired, action{action: ActionUpdate, object: "trustcenternda"})
	}

	key := "templateFiles"
	// get the file from the context, if it exists
	files, _ := objects.FilesFromContextWithKey(ctx, key)
	if len(files) != 1 {
		return &model.TrustCenterNDAUpdatePayload{
			Template: templates[0],
		}, nil
	}

	// Parse the template
	tmpl, err := template.New("nda").Parse(trustCenterNDATemplate)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	// Define the data to be used in the template
	data := struct {
		TrustCenterID string
		NDAFileID     string
	}{
		TrustCenterID: templates[0].TrustCenterID,
		NDAFileID:     files[0].ID,
	}

	// Create a bytes.Buffer to capture the output
	var buf bytes.Buffer

	// Execute the template, writing the output to the buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	// Get the output as a string from the buffer
	outputString := buf.String()
	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(templates[0]).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	return &model.TrustCenterNDAUpdatePayload{
		Template: updatedTmpl,
	}, nil
}
