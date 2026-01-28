//go:build test

package handlers_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/resend/resend-go/v2"
	"github.com/stretchr/testify/assert"

	"github.com/theopenlane/core/common/enums"
	"github.com/theopenlane/core/internal/ent/generated/privacy"
)

// testResendWebhookEvent is a test-only struct mirroring the Resend webhook event structure
// used to construct test payloads for webhook handler tests.
type testResendWebhookEvent struct {
	// Type is the event type, e.g. "email.sent".
	Type string `json:"type"`
	// Data contains the event payload.
	Data struct {
		// EmailID is the unique identifier for the email.
		EmailID string `json:"email_id"`
		// CreatedAt is the timestamp when the event occurred.
		CreatedAt string `json:"created_at"`
		// Tags are the custom tags attached to the email.
		Tags []struct {
			// Name is the tag key.
			Name string `json:"name"`
			// Value is the tag value.
			Value string `json:"value"`
		} `json:"tags"`
	} `json:"data"`
}

// signResendWebhook generates a valid HMAC-SHA256 signature for a Resend webhook payload.
// It uses the Svix signature format expected by the Resend SDK.
func signResendWebhook(secret, id, timestamp string, payload []byte) (string, error) {
	trimmed := strings.TrimPrefix(secret, "whsec_")
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return "", err
	}

	signedContent := fmt.Sprintf("%s.%s.%s", id, timestamp, string(payload))
	mac := hmac.New(sha256.New, decoded)
	_, _ = mac.Write([]byte(signedContent))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return "v1," + signature, nil
}

// TestResendWebhookHandlerEmailSent tests processing of email.sent events.
func (suite *HandlerTestSuite) TestResendWebhookHandlerEmailSent() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookEmailSent"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Email Sent Template").
		SetDescription("Template for email sent test").
		SetJsonconfig(map[string]any{"title": "Email Sent Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Sent Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Sent Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("emailsent@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{Type: resend.EventEmailSent}
	payload.Data.EmailID = "email_sent_001"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_sent_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.Equal(t, enums.AssessmentResponseStatusSent, updatedResp.Status)
	assert.NotNil(t, updatedResp.LastEmailEventAt)

	updatedTarget, err := suite.db.CampaignTarget.Get(ctx, target.ID)
	assert.NoError(t, err)
	assert.Equal(t, enums.AssessmentResponseStatusSent, updatedTarget.Status)
	assert.NotNil(t, updatedTarget.SentAt)
}

// TestResendWebhookHandlerEmailDelivered tests processing of email.delivered events.
func (suite *HandlerTestSuite) TestResendWebhookHandlerEmailDelivered() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookEmailDelivered"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Email Delivered Template").
		SetDescription("Template for email delivered test").
		SetJsonconfig(map[string]any{"title": "Email Delivered Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Delivered Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Delivered Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("emaildelivered@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{Type: resend.EventEmailDelivered}
	payload.Data.EmailID = "email_delivered_001"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_delivered_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedResp.LastEmailEventAt)
	assert.False(t, updatedResp.EmailDeliveredAt.IsZero(), "EmailDeliveredAt should be set")
}

// TestResendWebhookHandlerEmailOpened tests processing of email.opened events.
func (suite *HandlerTestSuite) TestResendWebhookHandlerEmailOpened() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookEmailOpened"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Email Opened Template").
		SetDescription("Template for email opened test").
		SetJsonconfig(map[string]any{"title": "Email Opened Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Opened Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Opened Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("emailopened@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{Type: resend.EventEmailOpened}
	payload.Data.EmailID = "email_opened_001"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_opened_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedResp.LastEmailEventAt)
	assert.False(t, updatedResp.EmailOpenedAt.IsZero(), "EmailOpenedAt should be set")
	assert.Equal(t, 1, updatedResp.EmailOpenCount, "EmailOpenCount should be incremented")
}

// TestResendWebhookHandlerEmailClicked tests processing of email.clicked events.
func (suite *HandlerTestSuite) TestResendWebhookHandlerEmailClicked() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookEmailClicked"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Email Clicked Template").
		SetDescription("Template for email clicked test").
		SetJsonconfig(map[string]any{"title": "Email Clicked Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Clicked Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Clicked Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("emailclicked@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{Type: resend.EventEmailClicked}
	payload.Data.EmailID = "email_clicked_001"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_clicked_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedResp.LastEmailEventAt)
	assert.False(t, updatedResp.EmailClickedAt.IsZero(), "EmailClickedAt should be set")
	assert.Equal(t, 1, updatedResp.EmailClickCount, "EmailClickCount should be incremented")
}

// TestResendWebhookHandlerEmailBounced tests processing of email.bounced events.
func (suite *HandlerTestSuite) TestResendWebhookHandlerEmailBounced() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookEmailBounced"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Email Bounced Template").
		SetDescription("Template for email bounced test").
		SetJsonconfig(map[string]any{"title": "Email Bounced Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Bounced Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Email Bounced Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("emailbounced@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	emailID := "email_bounced_001"
	payload := testResendWebhookEvent{Type: resend.EventEmailBounced}
	payload.Data.EmailID = emailID
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_bounced_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedResp.EmailMetadata, "EmailMetadata should be set")
	assert.Equal(t, resend.EventEmailBounced, updatedResp.EmailMetadata["resend_event"])
	assert.Equal(t, emailID, updatedResp.EmailMetadata["resend_email_id"])
	assert.Equal(t, eventTime, updatedResp.EmailMetadata["resend_timestamp"])

	updatedTarget, err := suite.db.CampaignTarget.Get(ctx, target.ID)
	assert.NoError(t, err)
	assert.NotNil(t, updatedTarget.Metadata, "campaign target Metadata should be set")
	assert.Equal(t, resend.EventEmailBounced, updatedTarget.Metadata["resend_event"])
}

// TestResendWebhookHandlerIsTestTag tests that when is_test tag is set to true,
// campaign target updates are skipped but assessment response updates still occur.
func (suite *HandlerTestSuite) TestResendWebhookHandlerIsTestTag() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookIsTestTag"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("IsTest Template").
		SetDescription("Template for is_test tag test").
		SetJsonconfig(map[string]any{"title": "IsTest Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("IsTest Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("IsTest Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("istest@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	// capture initial target state
	initialTarget, err := suite.db.CampaignTarget.Get(ctx, target.ID)
	assert.NoError(t, err)
	initialTargetStatus := initialTarget.Status

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{
		Type: resend.EventEmailSent,
	}
	payload.Data.EmailID = "email_istest_001"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
		{Name: "is_test", Value: "true"},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	msgID := "msg_istest_001"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, msgID, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", msgID)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	// verify assessment response was updated
	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.Equal(t, enums.AssessmentResponseStatusSent, updatedResp.Status, "assessment response should be updated")
	assert.NotNil(t, updatedResp.LastEmailEventAt, "LastEmailEventAt should be set")

	// verify campaign target was NOT updated due to is_test tag
	updatedTarget, err := suite.db.CampaignTarget.Get(ctx, target.ID)
	assert.NoError(t, err)
	assert.Equal(t, initialTargetStatus, updatedTarget.Status, "campaign target status should not change with is_test=true")
	assert.Nil(t, updatedTarget.SentAt, "SentAt should not be set with is_test=true")
}

// TestResendWebhookHandlerSequentialEvents tests that multiple events for the same
// response correctly accumulate counts and update timestamps.
func (suite *HandlerTestSuite) TestResendWebhookHandlerSequentialEvents() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookSequential"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Sequential Template").
		SetDescription("Template for sequential events test").
		SetJsonconfig(map[string]any{"title": "Sequential Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Sequential Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Sequential Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("sequential@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	// helper to send webhook event
	sendEvent := func(eventType, emailID, msgID string) {
		eventTime := time.Now().UTC().Format(time.RFC3339Nano)
		payload := testResendWebhookEvent{
			Type: eventType,
		}
		payload.Data.EmailID = emailID
		payload.Data.CreatedAt = eventTime
		payload.Data.Tags = []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}{
			{Name: "assessment_response_id", Value: response.ID},
			{Name: "campaign_target_id", Value: target.ID},
		}

		body, err := json.Marshal(payload)
		assert.NoError(t, err)

		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		signature, err := signResendWebhook(secret, msgID, timestamp, body)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
		req.Header.Set("svix-id", msgID)
		req.Header.Set("svix-timestamp", timestamp)
		req.Header.Set("svix-signature", signature)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)

		res := recorder.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	}

	// simulate realistic email lifecycle: sent -> delivered -> opened -> opened again -> clicked
	sendEvent(resend.EventEmailSent, "email_seq_001", "msg_seq_sent")
	sendEvent(resend.EventEmailDelivered, "email_seq_001", "msg_seq_delivered")
	sendEvent(resend.EventEmailOpened, "email_seq_001", "msg_seq_opened_1")
	sendEvent(resend.EventEmailOpened, "email_seq_001", "msg_seq_opened_2")
	sendEvent(resend.EventEmailOpened, "email_seq_001", "msg_seq_opened_3")
	sendEvent(resend.EventEmailClicked, "email_seq_001", "msg_seq_clicked_1")
	sendEvent(resend.EventEmailClicked, "email_seq_001", "msg_seq_clicked_2")

	// verify final state of assessment response
	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)

	assert.Equal(t, enums.AssessmentResponseStatusSent, updatedResp.Status, "status should be Sent")
	assert.NotNil(t, updatedResp.LastEmailEventAt, "LastEmailEventAt should be set")
	assert.False(t, updatedResp.EmailDeliveredAt.IsZero(), "EmailDeliveredAt should be set")
	assert.False(t, updatedResp.EmailOpenedAt.IsZero(), "EmailOpenedAt should be set")
	assert.False(t, updatedResp.EmailClickedAt.IsZero(), "EmailClickedAt should be set")
	assert.Equal(t, 3, updatedResp.EmailOpenCount, "EmailOpenCount should be 3 after 3 open events")
	assert.Equal(t, 2, updatedResp.EmailClickCount, "EmailClickCount should be 2 after 2 click events")

	// verify campaign target state
	updatedTarget, err := suite.db.CampaignTarget.Get(ctx, target.ID)
	assert.NoError(t, err)
	assert.Equal(t, enums.AssessmentResponseStatusSent, updatedTarget.Status, "campaign target status should be Sent")
	assert.NotNil(t, updatedTarget.SentAt, "SentAt should be set")
}

// TestResendWebhookHandlerInvalidSignature tests that webhook requests with an invalid
// HMAC signature are rejected with a 400 Bad Request status.
func (suite *HandlerTestSuite) TestResendWebhookHandlerInvalidSignature() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookInvalidSignature"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	payload := testResendWebhookEvent{
		Type: resend.EventEmailSent,
	}
	payload.Data.EmailID = "email_123"
	payload.Data.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	id := "msg_invalid"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook("whsec_"+base64.StdEncoding.EncodeToString([]byte("wrong-secret")), id, timestamp, body)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-id", id)
	req.Header.Set("svix-timestamp", timestamp)
	req.Header.Set("svix-signature", signature)

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

// TestResendWebhookHandlerIdempotent tests that duplicate webhook events with the same
// event ID are properly deduplicated and counters are only incremented once.
func (suite *HandlerTestSuite) TestResendWebhookHandlerIdempotent() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookIdempotent"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	ctx := privacy.DecisionContext(testUser1.UserCtx, privacy.Allow)

	template, err := suite.db.Template.Create().
		SetName("Campaign Template Idempotent").
		SetDescription("Campaign template").
		SetJsonconfig(map[string]any{"title": "Webhook Test"}).
		Save(ctx)
	assert.NoError(t, err)

	assessment, err := suite.db.Assessment.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Webhook Assessment").
		SetTemplateID(template.ID).
		Save(ctx)
	assert.NoError(t, err)

	campaign, err := suite.db.Campaign.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetName("Webhook Campaign").
		SetCampaignType(enums.CampaignTypeQuestionnaire).
		SetAssessmentID(assessment.ID).
		SetStatus(enums.CampaignStatusActive).
		SetIsActive(true).
		Save(ctx)
	assert.NoError(t, err)

	target, err := suite.db.CampaignTarget.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetCampaignID(campaign.ID).
		SetEmail("webhook.idempotent@example.com").
		Save(ctx)
	assert.NoError(t, err)

	response, err := suite.db.AssessmentResponse.Create().
		SetOwnerID(testUser1.OrganizationID).
		SetAssessmentID(assessment.ID).
		SetCampaignID(campaign.ID).
		SetEmail(target.Email).
		Save(ctx)
	assert.NoError(t, err)

	eventTime := time.Now().UTC().Format(time.RFC3339Nano)
	payload := testResendWebhookEvent{
		Type: resend.EventEmailOpened,
	}
	payload.Data.EmailID = "email_456"
	payload.Data.CreatedAt = eventTime
	payload.Data.Tags = []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}{
		{Name: "assessment_response_id", Value: response.ID},
		{Name: "campaign_target_id", Value: target.ID},
	}

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	id := "msg_idempotent"
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	signature, err := signResendWebhook(secret, id, timestamp, body)
	assert.NoError(t, err)

	makeRequest := func() *http.Response {
		req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
		req.Header.Set("svix-id", id)
		req.Header.Set("svix-timestamp", timestamp)
		req.Header.Set("svix-signature", signature)

		recorder := httptest.NewRecorder()
		suite.e.ServeHTTP(recorder, req)
		res := recorder.Result()
		return res
	}

	res := makeRequest()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	res = makeRequest()
	defer res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	updatedResp, err := suite.db.AssessmentResponse.Get(ctx, response.ID)
	assert.NoError(t, err)
	assert.Equal(t, 1, updatedResp.EmailOpenCount)
}

// TestResendWebhookHandlerMissingEventID tests that webhook requests missing the
// svix-id header are rejected with a 400 Bad Request status.
func (suite *HandlerTestSuite) TestResendWebhookHandlerMissingEventID() {
	t := suite.T()

	operation := openapi3.NewOperation()
	operation.OperationID = "ResendWebhookMissingEventID"
	suite.registerTestHandler(http.MethodPost, "/resend/webhook", operation, suite.h.ResendWebhookHandler)

	secret := "whsec_" + base64.StdEncoding.EncodeToString([]byte("resend-webhook-secret"))
	suite.h.CampaignWebhook.Enabled = true
	suite.h.CampaignWebhook.ResendSecret = secret

	payload := testResendWebhookEvent{
		Type: resend.EventEmailSent,
	}
	payload.Data.EmailID = "email_missing_id"
	payload.Data.CreatedAt = time.Now().UTC().Format(time.RFC3339Nano)

	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/resend/webhook", bytes.NewReader(body))
	req.Header.Set("svix-timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	req.Header.Set("svix-signature", "invalid")

	recorder := httptest.NewRecorder()
	suite.e.ServeHTTP(recorder, req)

	res := recorder.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}
