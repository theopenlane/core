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

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqljson"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/theopenlane/core/internal/graphapi/model"
	"github.com/theopenlane/core/pkg/enums"
	"github.com/theopenlane/core/pkg/logx"
	"github.com/theopenlane/core/pkg/objects"
	"github.com/theopenlane/core/pkg/olauth"
	"github.com/theopenlane/emailtemplates"
	"github.com/theopenlane/ent/generated"
	"github.com/theopenlane/ent/generated/documentdata"
	"github.com/theopenlane/ent/generated/privacy"
	gentemplate "github.com/theopenlane/ent/generated/template"
	"github.com/theopenlane/ent/privacy/rule"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/tokens"
	"github.com/theopenlane/riverboat/pkg/jobs"
	"github.com/theopenlane/utils/contextx"
	"github.com/theopenlane/utils/rout"
)

//go:embed trustcenternda.json.tpl
var trustCenterNDATemplate string

var (
	errTrustCenterOwnerNotFound = errors.New("trust center owner not found")
	errOneNDAOnly               = errors.New("one NDA file is required")
)

func createTrustCenterNDA(ctx context.Context, input model.CreateTrustCenterNDAInput) (*model.TrustCenterNDACreatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	trustCenter, err := txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	// set the organization in the auth context if its not done for us
	if err := setOrganizationInAuthContext(ctx, &trustCenter.OwnerID); err != nil {
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
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "template"})
	}

	// Parse the template
	tmpl, err := template.New("nda").Parse(trustCenterNDATemplate)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	key := "templateFiles"

	// get the file from the context, if it exists
	files, _ := objects.FilesFromContextWithKey(ctx, key)

	if len(files) != 1 {
		return nil, parseRequestError(ctx, errOneNDAOnly, action{action: ActionCreate, object: "trustcenternda"})
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
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	// Get the output as a string from the buffer
	outputString := buf.String()

	var outputInterface map[string]interface{}
	if err := json.Unmarshal([]byte(outputString), &outputInterface); err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(templateObj).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	return &model.TrustCenterNDACreatePayload{
		Template: updatedTmpl,
	}, nil
}

func updateTrustCenterNDA(ctx context.Context, id string) (*model.TrustCenterNDAUpdatePayload, error) {
	txnCtx := withTransactionalMutation(ctx)

	ndaTemplate, err := txnCtx.Template.Query().Where(gentemplate.TrustCenterIDEQ(id)).Only(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionUpdate, object: "trustcenternda"})
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
		return nil, parseRequestError(ctx, err, action{action: ActionUpdate, object: "trustcenternda"})
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
		return nil, parseRequestError(ctx, err, action{action: ActionUpdate, object: "trustcenternda"})
	}

	updatedTmpl, err := txnCtx.Template.UpdateOne(ndaTemplate).SetJsonconfig(outputInterface).Save(ctx)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionUpdate, object: "trustcenternda"})
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
			return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
		}

		if !admin {
			return nil, rout.ErrPermissionDenied
		}

		anonymousUser = &auth.AnonymousTrustCenterUser{
			SubjectID:          fmt.Sprintf("%s%s", olauth.AnonTrustcenterJWTPrefix, uuid.New().String()),
			SubjectName:        "Anonymous User",
			AuthenticationType: auth.JWTAuthentication,
			TrustCenterID:      input.TrustCenterID,
		}
	}

	txnCtx := withTransactionalMutation(ctx)

	trustCenter, err := txnCtx.TrustCenter.Get(ctx, input.TrustCenterID)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	allowCtx := privacy.DecisionContext(ctx, privacy.Allow)

	trustCenterOwner, err := txnCtx.Organization.Get(allowCtx, trustCenter.OwnerID)
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	if trustCenterOwner == nil {
		return nil, parseRequestError(ctx, errTrustCenterOwnerNotFound, action{action: ActionCreate, object: "trustcenterndaemail"})
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
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
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

	// Check if the user has already signed the NDA.
	// If they have, we'll send them an email with a new auth token
	ndaTemplate, err := txnCtx.Template.Query().Where(
		gentemplate.And(
			gentemplate.TrustCenterIDEQ(trustCenter.ID),
			gentemplate.KindEQ(enums.TemplateKindTrustCenterNda),
		),
	).Only(allowCtx)
	if err != nil && !generated.IsNotFound(err) {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	if ndaTemplate != nil {
		// Check if there's a document_data record for this user and template
		count, err := txnCtx.DocumentData.Query().Where(
			documentdata.And(
				documentdata.TemplateIDEQ(ndaTemplate.ID),
				func(s *sql.Selector) {
					s.Where(
						sqljson.ValueEQ(documentdata.FieldData, anonymousUser.SubjectEmail, sqljson.DotPath("signatory_info.email")),
					)
				},
				func(s *sql.Selector) {
					s.Where(
						sqljson.ValueEQ(documentdata.FieldData, anonymousUser.SubjectID, sqljson.DotPath("signature_metadata.user_id")),
					)
				},
			),
		).Count(allowCtx)
		if err != nil {
			return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
		}

		if count > 0 {
			// send the new link email
			email, err := txnCtx.Emailer.NewTrustCenterAuthEmail(emailtemplates.Recipient{
				Email: input.Email,
			}, access, emailtemplates.TrustCenterAuthData{
				OrganizationName: orgName,
				TrustCenterURL:   trustCenterURL.String(),
			})
			if err != nil {
				return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
			}
			// Send the email via job queue
			if _, err := r.db.Job.Insert(ctx, jobs.EmailArgs{
				Message: *email,
			}, nil); err != nil {
				return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
			}

			return &model.SendTrustCenterNDAEmailPayload{
				Success: true,
			}, nil
		}
	}

	trustCenterURL.Path += "/sign-nda"

	email, err := txnCtx.Emailer.NewTrustCenterNDARequestEmail(emailtemplates.Recipient{
		Email: input.Email,
	}, access, emailtemplates.TrustCenterNDARequestData{
		OrganizationName: orgName,
		TrustCenterURL:   trustCenterURL.String(),
	})
	if err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	// Send the email via job queue
	if _, err := r.db.Job.Insert(ctx, jobs.EmailArgs{
		Message: *email,
	}, nil); err != nil {
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenterndaemail"})
	}

	return &model.SendTrustCenterNDAEmailPayload{
		Success: true,
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
		return nil, parseRequestError(ctx, err, action{action: ActionCreate, object: "trustcenternda"})
	}

	return &model.SubmitTrustCenterNDAResponsePayload{
		DocumentData: res,
	}, nil
}
