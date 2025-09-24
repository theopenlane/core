package graphapi

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/theopenlane/core/internal/ent/generated"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
	gentemplate "github.com/theopenlane/core/internal/ent/generated/template"
	"github.com/theopenlane/core/internal/ent/privacy/rule"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/rout"
)

//go:embed trustcenternda.json.tpl
var trustCenterNDATemplate string

var (
	errOneNDAOnly               = errors.New("one NDA file is required")
	errTrustCenterOwnerNotFound = errors.New("trust center owner not found")
)

func createTrustCenterNDA(ctx context.Context, input model.CreateTrustCenterNDAInput) (*model.TrustCenterNDACreatePayload, error) {
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

func updateTrustCenterNDA(ctx context.Context, id string) (*model.TrustCenterNDAUpdatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)
	ndaTemplate, err := txnCtx.Template.Query().Where(gentemplate.TrustCenterIDEQ(id)).Only(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
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
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
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
		panic(err)
	}

	// Get the output as a string from the buffer
	outputString := buf.String()
	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(ndaTemplate).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	return &model.TrustCenterNDAUpdatePayload{
		Template: updatedTmpl,
	}, nil
}

func sendTrustCenterNDAEmail(ctx context.Context, input model.SendTrustCenterNDAInput, r *mutationResolver) (*model.SendTrustCenterNDAEmailPayload, error) {
	var anonymousUser *auth.AnonymousTrustCenterUser
	if anon, ok := auth.AnonymousTrustCenterUserFromContext(ctx); ok {
		if anon.TrustCenterID != input.TrustCenterID {
			return nil, rout.ErrPermissionDenied
		}
		anonymousUser = anon

	} else {
		// allow for system admins to also send the email
		admin, err := rule.CheckIsSystemAdminWithContext(ctx)
		if err != nil {
			return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
		}

		if !admin {
			return nil, rout.ErrPermissionDenied
		}
		anonymousUser = &auth.AnonymousTrustCenterUser{
			SubjectID:          fmt.Sprintf("anon_%s", uuid.New().String()),
			SubjectName:        "Anonymous User",
			AuthenticationType: auth.JWTAuthentication,
			TrustCenterID:      input.TrustCenterID,
		}
	}

	txnCtx := withTransactionalMutation(ctx)

	trustCenter, err := txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)
	trustCenterOwner, err := txnCtx.Organization.Get(allowCtx, trustCenter.OwnerID)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	if trustCenterOwner == nil {
		return nil, parseRequestError(errTrustCenterOwnerNotFound, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	orgName := trustCenterOwner.Name

	// create new claims for the user
	newClaims := &tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject: anonymousUser.SubjectID,
		},
		UserID:        anonymousUser.SubjectID,
		OrgID:         trustCenter.OwnerID,
		TrustCenterID: anonymousUser.TrustCenterID,
		Email:         input.Email,
	}

	// create a new token pair for the user
	access, _, err := r.db.TokenManager.CreateTokenPair(newClaims)
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	trustCenterURL := url.URL{
		Scheme: "https",
	}
	if trustCenter.Edges.CustomDomain != nil {
		trustCenterURL.Host = trustCenter.Edges.CustomDomain.CnameRecord
	} else {
		trustCenterURL.Host = r.defaultTrustCenterDomain
		trustCenterURL.Path = "/" + trustCenter.Slug
	}

	trustCenterURL.Path += "/sign-nda"

	email, err := txnCtx.Emailer.NewTrustCenterNDARequestEmail(emailtemplates.Recipient{
		Email: input.Email,
	}, access, emailtemplates.TrustCenterNDARequestData{
		OrganizationName: orgName,
		TrustCenterURL:   trustCenterURL.String(),
	})
	if err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	// Send the email via job queue
	if _, err := r.db.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		return nil, parseRequestError(err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	return &model.SendTrustCenterNDAEmailPayload{
		Success: true,
	}, nil
}
