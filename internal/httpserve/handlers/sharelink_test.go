package handlers_test

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/lo"
	"github.com/theopenlane/httpsling"
	"github.com/theopenlane/iam/auth"
	"github.com/theopenlane/iam/fgax"
	"github.com/theopenlane/iam/tokens"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	models "github.com/theopenlane/core/common/openapi"
	"github.com/theopenlane/core/common/storagetypes"
	ent "github.com/theopenlane/core/v2/internal/ent/generated"
	"github.com/theopenlane/core/v2/internal/ent/generated/file"
	"github.com/theopenlane/core/v2/internal/ent/generated/organizationsetting"
	"github.com/theopenlane/core/v2/internal/ent/generated/sharelinkaccessevent"
	authmiddleware "github.com/theopenlane/core/v2/pkg/middleware/auth"
	"github.com/theopenlane/core/v2/pkg/middleware/ratelimit"
	"github.com/theopenlane/core/v2/pkg/objects/storage/proxy"
)

const (
	shareTestSecret         = "forwarder-secret-for-tests"
	shareTestRecipient      = "bob@example.com"
	shareTestOtherRecipient = "carol@example.com"
	shareTestCode           = "123456"
	shareTestWrongCode      = "000000"
	shareTestWrongToken     = "not-the-share-token"
	shareTestUnknownID      = "01HZZUNKNOWNSHARE0000000000"
	shareTestIP             = "203.0.113.10"
	shareTestUserAgent      = "test-agent"
	shareTestBaseURL        = "https://files.example.com"
	shareTestCodeMaxAttempt = 5
	shareTestCodeLifetime   = 10 * time.Minute
	shareTestLinkLifetime   = time.Hour
	shareTestExpirySettle   = 2 * time.Second
	shareTestShortExpiry    = 1500 * time.Millisecond
)

// shareLinkCodePattern finds the six digit verification code in an email body
var shareLinkCodePattern = regexp.MustCompile(`\b\d{6}\b`)

// shareFixture is an organization with external sharing enabled and two stored R2 files
type shareFixture struct {
	ctx    context.Context
	userID string
	orgID  string
	files  []*ent.File
}

// seedShareFixture seeds a user whose personal organization allows external sharing and owns two R2 files
func (suite *HandlerTestSuite) seedShareFixture() shareFixture {
	t := suite.T()

	user := suite.userBuilder(context.Background())
	ctx := auth.WithInternalOperationContext(auth.NewTestContextWithOrgID(user.ID, user.PersonalOrgID))

	err := suite.db.OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(user.PersonalOrgID)).
		SetExternalSharingEnabled(true).
		Exec(ctx)
	assert.NilError(t, err)

	files := make([]*ent.File, 0, 2)
	for _, name := range []string{"report.pdf", "appendix.pdf"} {
		created, err := suite.db.File.Create().
			AddOrganizationIDs(user.PersonalOrgID).
			SetProvidedFileName(name).
			SetProvidedFileExtension("pdf").
			SetDetectedContentType("application/pdf").
			SetStorageProvider(string(storagetypes.R2Provider)).
			SetStorageVolume("openlane-prod-files").
			SetStoragePath("org/object/" + name).
			Save(ctx)
		assert.NilError(t, err)

		files = append(files, created)
	}

	return shareFixture{ctx: ctx, userID: user.ID, orgID: user.PersonalOrgID, files: files}
}

// shareLinkAccessEvents returns the access event rows written for the share link
func (suite *HandlerTestSuite) shareLinkAccessEvents(fx shareFixture, linkID string) []*ent.ShareLinkAccessEvent {
	events, err := suite.db.ShareLinkAccessEvent.Query().
		Where(sharelinkaccessevent.ShareLinkID(linkID)).
		All(fx.ctx)
	assert.NilError(suite.T(), err)

	return events
}

// createShareLink creates a share link over the fixture's first file
func (suite *HandlerTestSuite) createShareLink(fx shareFixture, recipients []string, maxUses int, expiresAt time.Time) *ent.ShareLink {
	create := suite.db.ShareLink.Create().
		SetName("audit evidence").
		SetOwnerID(fx.orgID).
		SetMaxUses(maxUses).
		SetExpiresAt(expiresAt).
		AddFileIDs(fx.files[0].ID)

	if len(recipients) > 0 {
		create.SetRecipientEmails(recipients)
	}

	link, err := create.Save(fx.ctx)
	assert.NilError(suite.T(), err)

	return link
}

// shareCodeHash is the digest the code endpoint stores for a verification code issued to the recipient
func shareCodeHash(recipient, code string) []byte {
	sum := sha256.Sum256([]byte(proxy.RecipientHash(recipient) + ":" + code))

	return sum[:]
}

// seedShareCode stores a verification code at rest on the share link the way the code endpoint does for the recipient
func (suite *HandlerTestSuite) seedShareCode(fx shareFixture, link *ent.ShareLink, recipient, code string, expiresAt time.Time) {
	err := suite.db.ShareLink.UpdateOneID(link.ID).
		SetCodeHash(shareCodeHash(recipient, code)).
		SetCodeExpiresAt(expiresAt).
		SetCodeAttempts(0).
		Exec(fx.ctx)
	assert.NilError(suite.T(), err)
}

// setExternalSharing sets whether the fixture's organization allows external sharing
func (suite *HandlerTestSuite) setExternalSharing(fx shareFixture, enabled bool) {
	err := suite.db.OrganizationSetting.Update().
		Where(organizationsetting.OrganizationID(fx.orgID)).
		SetExternalSharingEnabled(enabled).
		Exec(fx.ctx)
	assert.NilError(suite.T(), err)
}

// fileViewTuple is the tuple upload writes so the members of the fixture's organization can view the file
func fileViewTuple(fx shareFixture, fileID string) fgax.TupleKey {
	return fgax.GetTupleKey(fgax.TupleRequest{
		SubjectID:       fx.orgID,
		SubjectType:     "organization",
		SubjectRelation: fgax.MemberRelation,
		ObjectID:        fileID,
		ObjectType:      "file",
		Relation:        fgax.CanView,
	})
}

// requestShareCode asks the code endpoint for a verification code for the email
func (suite *HandlerTestSuite) requestShareCode(link *ent.ShareLink, email string) *httptest.ResponseRecorder {
	return suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token, Email: email})
}

// postShare posts a JSON body to a share link endpoint, presenting the forwarder secret when one is given
func (suite *HandlerTestSuite) postShare(path, secret string, body any) *httptest.ResponseRecorder {
	raw, err := json.Marshal(body)
	assert.NilError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(string(raw)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)

	if secret != "" {
		req.Header.Set(ratelimit.ForwarderSecretHeader, secret)
	}

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	return recorder
}

// registerShareRoutes registers the share link endpoints on the test router and configures the forwarder secret and
// the auth options session bound downloads authenticate with
func (suite *HandlerTestSuite) registerShareRoutes() {
	suite.h.ForwarderSecret = shareTestSecret

	keys, err := suite.db.TokenManager.Keys()
	assert.NilError(suite.T(), err)

	conf := authmiddleware.NewAuthOptions(
		authmiddleware.WithDBClient(suite.db),
		authmiddleware.WithValidator(tokens.NewJWKSValidator(keys, testTokenIssuer, testTokenIssuer)),
	)
	suite.h.AuthOptions = &conf

	suite.registerTestHandler(http.MethodPost, "share/:id/resolve", suite.h.ShareLinkResolveHandler)
	suite.registerTestHandler(http.MethodPost, "share/:id/code", suite.h.ShareLinkCodeHandler)
	suite.registerTestHandler(http.MethodPost, "share/:id/verify", suite.h.ShareLinkVerifyHandler)
	suite.registerTestHandler(http.MethodPost, "share/:id/consume", suite.h.ShareLinkConsumeHandler)
}

// sessionToken mints a session access token for the user in the organization
func (suite *HandlerTestSuite) sessionToken(userID, orgID string) string {
	access, _, err := suite.h.TokenManager.CreateTokenPair(&tokens.Claims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
		UserID:           userID,
		OrgID:            orgID,
	})
	assert.NilError(suite.T(), err)

	return access
}

// verifiedToken mints the share-verified token the verify endpoint returns for the recipient of the share link
func (suite *HandlerTestSuite) verifiedToken(linkID, recipient string) string {
	token, err := proxy.NewShareVerifiedToken(suite.h.TokenManager, suite.h.TokenManager.Config().Issuer, linkID, proxy.RecipientHash(recipient))
	assert.NilError(suite.T(), err)

	return token
}

// rawConsume posts a consume request without asserting, so it is safe to call from a goroutine; the bearer is
// forwarded as the Authorization header when one is given
func (suite *HandlerTestSuite) rawConsume(linkID string, in models.ShareLinkConsumeRequest, bearer string) *httptest.ResponseRecorder {
	in.IP = shareTestIP
	in.UserAgent = shareTestUserAgent

	raw, _ := json.Marshal(in)

	req := httptest.NewRequest(http.MethodPost, "/share/"+linkID+"/consume", strings.NewReader(string(raw)))
	req.Header.Set(httpsling.HeaderContentType, httpsling.ContentTypeJSONUTF8)
	req.Header.Set(ratelimit.ForwarderSecretHeader, shareTestSecret)

	if bearer != "" {
		req.Header.Set(httpsling.HeaderAuthorization, "Bearer "+bearer)
	}

	recorder := httptest.NewRecorder()

	suite.e.ServeHTTP(recorder, req)

	return recorder
}

// consumeWith posts a consume request and decodes the response
func (suite *HandlerTestSuite) consumeWith(linkID string, in models.ShareLinkConsumeRequest, bearer string) (int, models.ShareLinkConsumeResponse) {
	recorder := suite.rawConsume(linkID, in, bearer)

	var out models.ShareLinkConsumeResponse
	if recorder.Code == http.StatusOK {
		assert.NilError(suite.T(), json.Unmarshal(recorder.Body.Bytes(), &out))
	}

	return recorder.Code, out
}

// consume posts a consume request for the share link and file presenting its share token
func (suite *HandlerTestSuite) consume(link *ent.ShareLink, fileID string) (int, models.ShareLinkConsumeResponse) {
	return suite.consumeWith(link.ID, models.ShareLinkConsumeRequest{Token: link.Token, FileID: fileID}, "")
}

// resolve posts a resolve request for the share link presenting the token
func (suite *HandlerTestSuite) resolve(linkID, token string) *httptest.ResponseRecorder {
	return suite.postShare("/share/"+linkID+"/resolve", shareTestSecret, models.ShareLinkResolveRequest{Token: token})
}

func TestShareLinkForwarderSecretGate(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()
	link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))

	paths := map[string]any{
		"/share/" + link.ID + "/resolve": models.ShareLinkResolveRequest{Token: link.Token},
		"/share/" + link.ID + "/code":    models.ShareLinkCodeRequest{Token: link.Token, Email: shareTestRecipient},
		"/share/" + link.ID + "/verify":  models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode},
		"/share/" + link.ID + "/consume": models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID},
	}

	t.Run("missing or wrong secret", func(t *testing.T) {
		for path, body := range paths {
			assert.Check(t, is.Equal(http.StatusForbidden, suite.postShare(path, "", body).Code), path)
			assert.Check(t, is.Equal(http.StatusForbidden, suite.postShare(path, "not-the-secret", body).Code), path)
		}
	})

	t.Run("disabled without configured secret", func(t *testing.T) {
		suite.h.ForwarderSecret = ""

		for path, body := range paths {
			assert.Check(t, is.Equal(http.StatusNotFound, suite.postShare(path, shareTestSecret, body).Code), path)
		}
	})
}

func TestShareLinkCodeHandler(t *testing.T) {
	suite := suite.setupTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()

	t.Run("emails the code only to a bound recipient", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestOtherRecipient, "  Bob@Example.COM "}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.mockEmailSender().Reset()

		recorder := suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token, Email: "BOB@example.com"})
		assert.Assert(t, is.Equal(http.StatusNoContent, recorder.Code), recorder.Body.String())

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, len(stored.CodeHash) > 0)
		assert.Assert(t, stored.CodeExpiresAt != nil)
		assert.Check(t, stored.CodeExpiresAt.Sub(time.Now().Add(shareTestCodeLifetime)).Abs() < time.Minute)
		assert.Check(t, is.Equal(0, stored.CodeAttempts))

		suite.WaitForEvents()

		msgs := suite.mockEmailSender().Messages()
		assert.Assert(t, is.Len(msgs, 1))
		assert.Check(t, is.DeepEqual([]string{shareTestRecipient}, msgs[0].To))

		code := shareLinkCodePattern.FindString(msgs[0].Text + msgs[0].HTML)
		assert.Assert(t, code != "", "the email carries the six digit code")

		assert.Check(t, is.DeepEqual(shareCodeHash(shareTestRecipient, code), stored.CodeHash), "the stored digest covers the recipient the code was issued to")
	})

	t.Run("issues no new code while one is pending", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient, shareTestOtherRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.mockEmailSender().Reset()

		assert.Assert(t, is.Equal(http.StatusNoContent, suite.requestShareCode(link, shareTestRecipient).Code))
		suite.WaitForEvents()
		assert.Assert(t, is.Len(suite.mockEmailSender().Messages(), 1))

		issued, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Assert(t, len(issued.CodeHash) > 0)

		suite.mockEmailSender().Reset()

		// neither the same recipient nor another one gets a second code in the window
		for _, email := range []string{shareTestRecipient, shareTestOtherRecipient} {
			assert.Check(t, is.Equal(http.StatusNoContent, suite.requestShareCode(link, email).Code))
		}

		suite.WaitForEvents()
		assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 0))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(issued.CodeHash, stored.CodeHash))
		assert.Check(t, stored.CodeExpiresAt.Equal(*issued.CodeExpiresAt))
	})

	t.Run("issues a new code once the pending one has expired", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(-time.Minute))
		suite.mockEmailSender().Reset()

		assert.Assert(t, is.Equal(http.StatusNoContent, suite.requestShareCode(link, shareTestRecipient).Code))

		suite.WaitForEvents()

		msgs := suite.mockEmailSender().Messages()
		assert.Assert(t, is.Len(msgs, 1))

		code := shareLinkCodePattern.FindString(msgs[0].Text + msgs[0].HTML)

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(shareCodeHash(shareTestRecipient, code), stored.CodeHash))
		assert.Assert(t, stored.CodeExpiresAt != nil)
		assert.Check(t, stored.CodeExpiresAt.After(time.Now()))
	})

	t.Run("issues a new code after the pending one was verified", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))
		suite.mockEmailSender().Reset()

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

		assert.Assert(t, is.Equal(http.StatusNoContent, suite.requestShareCode(link, shareTestRecipient).Code))

		suite.WaitForEvents()
		assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 1))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, len(stored.CodeHash) > 0)
	})

	t.Run("silently ignores a non recipient", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.mockEmailSender().Reset()

		recorder := suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token, Email: "mallory@example.com"})
		assert.Check(t, is.Equal(http.StatusNoContent, recorder.Code))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Len(stored.CodeHash, 0))

		suite.WaitForEvents()
		assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 0))
	})

	t.Run("silently ignores a revoked share and an unknown share", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		assert.NilError(t, suite.db.ShareLink.UpdateOneID(link.ID).SetRevokedAt(time.Now()).Exec(fx.ctx))
		suite.mockEmailSender().Reset()

		recorder := suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token, Email: shareTestRecipient})
		assert.Check(t, is.Equal(http.StatusNoContent, recorder.Code))

		suite.WaitForEvents()
		assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 0))

		recorder = suite.postShare("/share/"+shareTestUnknownID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token, Email: shareTestRecipient})
		assert.Check(t, is.Equal(http.StatusNoContent, recorder.Code))
	})

	t.Run("rejects a missing email", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: link.Token})
		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
	})

	t.Run("silently ignores a recipient presenting the wrong share token", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.mockEmailSender().Reset()

		recorder := suite.postShare("/share/"+link.ID+"/code", shareTestSecret, models.ShareLinkCodeRequest{Token: shareTestWrongToken, Email: shareTestRecipient})
		assert.Check(t, is.Equal(http.StatusNoContent, recorder.Code))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Len(stored.CodeHash, 0))

		suite.WaitForEvents()
		assert.Check(t, is.Len(suite.mockEmailSender().Messages(), 0))
	})
}

func TestShareLinkVerifyHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()

	t.Run("returns a share verified token naming the matched recipient", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestOtherRecipient, shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: " BOB@example.com", Code: shareTestCode})
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

		var out models.ShareLinkVerifyResponse
		assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &out))

		shareID, recipient, err := proxy.VerifyShareVerifiedToken(suite.h.TokenManager, suite.h.TokenManager.Config().Issuer, out.Token)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(link.ID, shareID))
		assert.Check(t, is.Equal(proxy.RecipientHash(shareTestRecipient), recipient))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Len(stored.CodeHash, 0))
		assert.Check(t, is.Nil(stored.CodeExpiresAt))

		recorder = suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code), "the code is single use")
	})

	t.Run("counts failures and keeps the exhausted code until it expires", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))

		for attempt := 1; attempt <= shareTestCodeMaxAttempt; attempt++ {
			recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestWrongCode})
			assert.Assert(t, is.Equal(http.StatusForbidden, recorder.Code))

			stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(attempt, stored.CodeAttempts))
		}

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code), "the right code no longer verifies once the attempts are used")

		// the exhausted code is still pending, so a code request cannot replace it and reset the attempts
		assert.Check(t, is.Equal(http.StatusNoContent, suite.requestShareCode(link, shareTestRecipient).Code))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.DeepEqual(shareCodeHash(shareTestRecipient, shareTestCode), stored.CodeHash), "the exhausted code stays stored until it expires")
		assert.Assert(t, stored.CodeExpiresAt != nil)
		assert.Check(t, stored.CodeExpiresAt.After(time.Now()))
		assert.Check(t, is.Equal(shareTestCodeMaxAttempt, stored.CodeAttempts))
	})

	t.Run("a code only verifies with the recipient it was issued to", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient, shareTestOtherRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestOtherRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code), "another bound recipient cannot use the code")

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, stored.CodeAttempts))
		assert.Check(t, len(stored.CodeHash) > 0)

		recorder = suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())
	})

	t.Run("rejects the wrong email, an expired code, and an unknown share", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: "mallory@example.com", Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code))

		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(-time.Minute))

		recorder = suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code))

		recorder = suite.postShare("/share/"+shareTestUnknownID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: link.Token, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code))
	})

	t.Run("rejects the right code presented with the wrong share token", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
		suite.seedShareCode(fx, link, shareTestRecipient, shareTestCode, time.Now().Add(shareTestCodeLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/verify", shareTestSecret, models.ShareLinkVerifyRequest{Token: shareTestWrongToken, Email: shareTestRecipient, Code: shareTestCode})
		assert.Check(t, is.Equal(http.StatusForbidden, recorder.Code))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, stored.CodeAttempts), "a bad share token does not count against the code")
	})
}

func TestShareLinkConsumeHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()

	t.Run("allows until the use cap and records the access", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 2, time.Now().Add(shareTestLinkLifetime))

		for range 2 {
			code, out := suite.consume(link, fx.files[0].ID)
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, out.Allowed)
			assert.Check(t, is.Equal("", out.Reason))
			assert.Assert(t, out.File != nil)
			assert.Check(t, is.Equal(fx.files[0].StoragePath, out.File.Key))
			assert.Check(t, is.Equal(fx.files[0].DetectedContentType, out.File.ContentType))
			assert.Check(t, is.Equal(fx.files[0].ProvidedFileName, out.File.FileName))
		}

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(2, stored.UseCount))

		events := suite.shareLinkAccessEvents(fx, link.ID)
		assert.Assert(t, is.Len(events, 2), "one access event row per allowed download")

		for _, event := range events {
			assert.Check(t, is.Equal(fx.files[0].ID, event.FileID))
			assert.Check(t, is.Equal(shareTestIP, event.IPAddress))
			assert.Check(t, is.Equal(shareTestUserAgent, event.UserAgent))
			assert.Check(t, is.Equal(fx.orgID, event.OwnerID))
		}

		code, out := suite.consume(link, fx.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonExhausted, out.Reason))
		assert.Check(t, out.File == nil, "a denied download names no object")
	})

	t.Run("unlimited uses never exhaust", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 0, time.Now().Add(shareTestLinkLifetime))

		for range 3 {
			code, out := suite.consume(link, fx.files[0].ID)
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, out.Allowed)
		}
	})

	t.Run("refuses a file other than the shared one", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		code, out := suite.consume(link, fx.files[1].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonUnknownFile, out.Reason))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 0), "a denied download writes no access event")
	})

	t.Run("a soft deleted file costs no use", func(t *testing.T) {
		gone := suite.seedShareFixture()
		link := suite.createShareLink(gone, nil, 1, time.Now().Add(shareTestLinkLifetime))

		assert.NilError(t, suite.db.File.Update().Where(file.ID(gone.files[0].ID)).SetDeletedAt(time.Now()).Exec(gone.ctx))

		code, out := suite.consume(link, gone.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonUnknownFile, out.Reason))
		assert.Check(t, out.File == nil)

		stored, err := suite.db.ShareLink.Get(gone.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(gone, link.ID), 0))
	})

	t.Run("a cleared use cap is unlimited", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		assert.NilError(t, suite.db.ShareLink.UpdateOneID(link.ID).ClearMaxUses().Exec(fx.ctx))

		for range 3 {
			code, out := suite.consume(link, fx.files[0].ID)
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, out.Allowed, out.Reason)
		}

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(3, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 3))
		assert.Check(t, is.Equal(http.StatusOK, suite.resolve(link.ID, link.Token).Code), "a share without a cap is never exhausted")
	})

	t.Run("allows each file of a multi file share and records the requested file", func(t *testing.T) {
		link, err := suite.db.ShareLink.Create().
			SetName("audit evidence").
			SetOwnerID(fx.orgID).
			SetMaxUses(0).
			SetExpiresAt(time.Now().Add(shareTestLinkLifetime)).
			AddFileIDs(lo.Map(fx.files, func(f *ent.File, _ int) string { return f.ID })...).
			Save(fx.ctx)
		assert.NilError(t, err)

		for _, shared := range fx.files {
			code, out := suite.consume(link, shared.ID)
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, out.Allowed)
		}

		events := suite.shareLinkAccessEvents(fx, link.ID)
		assert.Assert(t, is.Len(events, len(fx.files)))
		assert.Check(t, is.DeepEqual(
			slices.Sorted(slices.Values(lo.Map(fx.files, func(f *ent.File, _ int) string { return f.ID }))),
			slices.Sorted(slices.Values(lo.Map(events, func(e *ent.ShareLinkAccessEvent, _ int) string { return e.FileID }))),
		))
	})

	t.Run("refuses a revoked share", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		assert.NilError(t, suite.db.ShareLink.UpdateOneID(link.ID).SetRevokedAt(time.Now()).Exec(fx.ctx))

		code, out := suite.consume(link, fx.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonRevoked, out.Reason))
	})

	t.Run("refuses an expired share", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestShortExpiry))

		time.Sleep(shareTestExpirySettle)

		code, out := suite.consume(link, fx.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonExpired, out.Reason))
	})

	t.Run("unknown share is not found", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		code, _ := suite.consumeWith(shareTestUnknownID, models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID}, "")
		assert.Check(t, is.Equal(http.StatusNotFound, code))
	})

	t.Run("a token that does not verify is not found and consumes nothing", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		other := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		for _, token := range []string{shareTestWrongToken, other.Token} {
			code, _ := suite.consumeWith(link.ID, models.ShareLinkConsumeRequest{Token: token, FileID: fx.files[0].ID}, "")
			assert.Check(t, is.Equal(http.StatusNotFound, code))
		}

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 0))
	})

	t.Run("rejects a missing file id", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		recorder := suite.postShare("/share/"+link.ID+"/consume", shareTestSecret, models.ShareLinkConsumeRequest{Token: link.Token})
		assert.Check(t, is.Equal(http.StatusBadRequest, recorder.Code))
	})
}

func TestShareLinkConsumeHandlerRecipientBound(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()
	link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))
	other := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))

	testCases := []struct {
		name          string
		verifiedToken string
	}{
		{name: "missing verified token"},
		{name: "not a token", verifiedToken: shareTestWrongToken},
		{name: "verified token of another share", verifiedToken: suite.verifiedToken(other.ID, shareTestRecipient)},
		{name: "verified token of an address the share is not bound to", verifiedToken: suite.verifiedToken(link.ID, shareTestOtherRecipient)},
		{name: "session token in place of a verified token", verifiedToken: suite.sessionToken(fx.userID, fx.orgID)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := suite.consumeWith(link.ID, models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID, VerifiedToken: tc.verifiedToken}, "")
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, !out.Allowed)
			assert.Check(t, is.Equal(models.ShareLinkConsumeReasonVerificationRequired, out.Reason))
			assert.Check(t, out.File == nil)
		})
	}

	stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0, stored.UseCount), "an unverified recipient does not consume a use")
	assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 0))

	t.Run("verified recipient downloads", func(t *testing.T) {
		code, out := suite.consumeWith(link.ID, models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID, VerifiedToken: suite.verifiedToken(link.ID, " BOB@example.com")}, "")
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, out.Allowed)
		assert.Assert(t, out.File != nil)
		assert.Check(t, is.Equal(fx.files[0].StoragePath, out.File.Key))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 1))
	})
}

func TestShareLinkConsumeHandlerRequireSession(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()
	stranger := suite.userBuilder(context.Background())

	// the creator reads the file under the normal privacy rules, which need the organization's modules
	suite.enableModules(fx.userID, fx.orgID, nil)

	// a file record created directly carries no access tuples, upload writes them
	viewTuple := fileViewTuple(fx, fx.files[0].ID)

	_, err := suite.db.Authz.WriteTupleKeys(fx.ctx, []fgax.TupleKey{viewTuple}, nil)
	assert.NilError(t, err)

	link, err := suite.db.ShareLink.Create().
		SetOwnerID(fx.orgID).
		SetMaxUses(0).
		SetRequireSession(true).
		SetExpiresAt(time.Now().Add(shareTestLinkLifetime)).
		AddFileIDs(fx.files[0].ID).
		Save(fx.ctx)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(fx.userID, link.CreatedBy))

	in := models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID}

	testCases := []struct {
		name           string
		bearer         string
		expectedReason string
	}{
		{name: "no credential", expectedReason: models.ShareLinkConsumeReasonSessionRequired},
		{name: "credential that does not verify", bearer: shareTestWrongToken, expectedReason: models.ShareLinkConsumeReasonSessionRequired},
		{name: "authenticated caller other than the creator", bearer: suite.sessionToken(stranger.ID, stranger.PersonalOrgID), expectedReason: models.ShareLinkConsumeReasonForbidden},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			code, out := suite.consumeWith(link.ID, in, tc.bearer)
			assert.Assert(t, is.Equal(http.StatusOK, code))
			assert.Check(t, !out.Allowed)
			assert.Check(t, is.Equal(tc.expectedReason, out.Reason))
			assert.Check(t, out.File == nil)
		})
	}

	stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(0, stored.UseCount), "a refused session does not consume a use")
	assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 0))

	t.Run("creator downloads", func(t *testing.T) {
		code, out := suite.consumeWith(link.ID, in, suite.sessionToken(fx.userID, fx.orgID))
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, out.Allowed, out.Reason)
		assert.Assert(t, out.File != nil)
		assert.Check(t, is.Equal(fx.files[0].StoragePath, out.File.Key))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 1))
	})

	t.Run("creator in another organization is forbidden", func(t *testing.T) {
		// the creator's session is scoped to an organization the file does not belong to
		code, out := suite.consumeWith(link.ID, in, suite.sessionToken(fx.userID, stranger.PersonalOrgID))
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonForbidden, out.Reason))

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, stored.UseCount))
	})

	t.Run("creator who can no longer read the file is forbidden", func(t *testing.T) {
		// the creator stays in the organization and only loses view access to the file
		_, err := suite.db.Authz.WriteTupleKeys(fx.ctx, nil, []fgax.TupleKey{viewTuple})
		assert.NilError(t, err)

		code, out := suite.consumeWith(link.ID, in, suite.sessionToken(fx.userID, fx.orgID))
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, !out.Allowed)
		assert.Check(t, is.Equal(models.ShareLinkConsumeReasonForbidden, out.Reason))
		assert.Check(t, out.File == nil)

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(1, stored.UseCount))
	})
}

func TestShareLinkResolveHandler(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()

	t.Run("describes the share without consuming a use", func(t *testing.T) {
		link := suite.createShareLink(fx, []string{shareTestRecipient}, 1, time.Now().Add(shareTestLinkLifetime))

		for range 2 {
			recorder := suite.resolve(link.ID, link.Token)
			assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

			var out models.ShareLinkResolveResponse
			assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &out))

			assert.Check(t, is.Equal("audit evidence", out.Name))
			assert.Check(t, out.Organization != "")
			assert.Check(t, out.ExpiresAt.Equal(link.ExpiresAt))
			assert.Check(t, out.LandingPage)
			assert.Check(t, out.RecipientBound)
			assert.Check(t, is.DeepEqual([]models.ShareLinkResolveFile{{ID: fx.files[0].ID, Name: fx.files[0].ProvidedFileName, ContentType: fx.files[0].DetectedContentType}}, out.Files))
		}

		// the worker reads these exact field names
		var raw map[string]any
		assert.NilError(t, json.Unmarshal(suite.resolve(link.ID, link.Token).Body.Bytes(), &raw))

		for _, field := range []string{"name", "organization", "expires_at", "landing_page", "recipient_bound", "files"} {
			_, ok := raw[field]
			assert.Check(t, ok, field)
		}

		expiresAt, ok := raw["expires_at"].(string)
		assert.Assert(t, ok)

		_, err := time.Parse(time.RFC3339, expiresAt)
		assert.Check(t, err)

		stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
		assert.NilError(t, err)
		assert.Check(t, is.Equal(0, stored.UseCount))
		assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), 0))
	})

	t.Run("direct single file share", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 0, time.Now().Add(shareTestLinkLifetime))

		recorder := suite.resolve(link.ID, link.Token)
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

		var out models.ShareLinkResolveResponse
		assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &out))
		assert.Check(t, !out.LandingPage)
		assert.Check(t, !out.RecipientBound)
		assert.Check(t, is.Len(out.Files, 1))
	})

	t.Run("capped single file share opens the landing page", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		recorder := suite.resolve(link.ID, link.Token)
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

		var out models.ShareLinkResolveResponse
		assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &out))
		assert.Check(t, out.LandingPage)
		assert.Check(t, !out.RecipientBound)
	})

	t.Run("unknown share and a token that does not verify are not found", func(t *testing.T) {
		link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		other := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

		assert.Check(t, is.Equal(http.StatusNotFound, suite.resolve(shareTestUnknownID, link.Token).Code))
		assert.Check(t, is.Equal(http.StatusNotFound, suite.resolve(link.ID, shareTestWrongToken).Code))
		assert.Check(t, is.Equal(http.StatusNotFound, suite.resolve(link.ID, other.Token).Code))
		assert.Check(t, is.Equal(http.StatusBadRequest, suite.resolve(link.ID, "").Code))
	})

	t.Run("gone with the reason", func(t *testing.T) {
		revoked := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		assert.NilError(t, suite.db.ShareLink.UpdateOneID(revoked.ID).SetRevokedAt(time.Now()).Exec(fx.ctx))

		exhausted := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
		code, out := suite.consume(exhausted, fx.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Assert(t, out.Allowed)

		expired := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestShortExpiry))

		direct := suite.createShareLink(fx, nil, 0, time.Now().Add(shareTestLinkLifetime))
		assert.NilError(t, suite.db.ShareLink.UpdateOneID(direct.ID).SetRevokedAt(time.Now()).Exec(fx.ctx))

		time.Sleep(shareTestExpirySettle)

		testCases := []struct {
			name                string
			link                *ent.ShareLink
			expectedReason      string
			expectedLandingPage bool
		}{
			{name: "revoked", link: revoked, expectedReason: models.ShareLinkConsumeReasonRevoked, expectedLandingPage: true},
			{name: "expired", link: expired, expectedReason: models.ShareLinkConsumeReasonExpired, expectedLandingPage: true},
			{name: "exhausted", link: exhausted, expectedReason: models.ShareLinkConsumeReasonExhausted, expectedLandingPage: true},
			{name: "revoked direct share", link: direct, expectedReason: models.ShareLinkConsumeReasonRevoked},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				recorder := suite.resolve(tc.link.ID, tc.link.Token)
				assert.Assert(t, is.Equal(http.StatusGone, recorder.Code), recorder.Body.String())

				var gone models.ShareLinkGoneResponse
				assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &gone))
				assert.Check(t, is.Equal(tc.expectedReason, gone.Reason))
				assert.Check(t, is.Equal(tc.expectedLandingPage, gone.LandingPage))

				// the worker reads the landing page flag by this exact field name
				var raw map[string]any
				assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &raw))
				assert.Check(t, is.Equal(tc.expectedLandingPage, raw["landing_page"]))
			})
		}

		assert.Check(t, is.Equal(http.StatusNotFound, suite.resolve(expired.ID, shareTestWrongToken).Code), "an expired share still hides a token that does not match")
	})
}

func TestShareLinkExternalSharingDisabled(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()
	capped := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))
	direct := suite.createShareLink(fx, nil, 0, time.Now().Add(shareTestLinkLifetime))

	suite.setExternalSharing(fx, false)

	testCases := []struct {
		name                string
		link                *ent.ShareLink
		expectedLandingPage bool
	}{
		{name: "capped share", link: capped, expectedLandingPage: true},
		{name: "direct share", link: direct},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := suite.resolve(tc.link.ID, tc.link.Token)
			assert.Assert(t, is.Equal(http.StatusGone, recorder.Code), recorder.Body.String())

			var gone models.ShareLinkGoneResponse
			assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &gone))
			assert.Check(t, is.Equal(models.ShareLinkConsumeReasonRevoked, gone.Reason))
			assert.Check(t, is.Equal(tc.expectedLandingPage, gone.LandingPage))

			// a file outside the share is reported as revoked too
			for _, shared := range fx.files {
				code, out := suite.consume(tc.link, shared.ID)
				assert.Assert(t, is.Equal(http.StatusOK, code))
				assert.Check(t, !out.Allowed)
				assert.Check(t, is.Equal(models.ShareLinkConsumeReasonRevoked, out.Reason))
				assert.Check(t, out.File == nil)
			}

			stored, err := suite.db.ShareLink.Get(fx.ctx, tc.link.ID)
			assert.NilError(t, err)
			assert.Check(t, is.Equal(0, stored.UseCount))
			assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, tc.link.ID), 0))
		})
	}

	t.Run("sharing turned back on serves the share again", func(t *testing.T) {
		suite.setExternalSharing(fx, true)

		assert.Check(t, is.Equal(http.StatusOK, suite.resolve(capped.ID, capped.Token).Code))

		code, out := suite.consume(capped, fx.files[0].ID)
		assert.Assert(t, is.Equal(http.StatusOK, code))
		assert.Check(t, out.Allowed, out.Reason)
	})
}

func TestShareLinkConsumeHandlerConcurrent(t *testing.T) {
	suite := suite.setupParallelTest(t)
	suite.registerShareRoutes()

	fx := suite.seedShareFixture()
	link := suite.createShareLink(fx, nil, 1, time.Now().Add(shareTestLinkLifetime))

	const concurrentDownloads = 2

	recorders := make([]*httptest.ResponseRecorder, concurrentDownloads)

	var (
		start sync.WaitGroup
		done  sync.WaitGroup
	)

	start.Add(1)

	for i := range concurrentDownloads {
		done.Add(1)

		go func(i int) {
			defer done.Done()

			start.Wait()

			recorders[i] = suite.rawConsume(link.ID, models.ShareLinkConsumeRequest{Token: link.Token, FileID: fx.files[0].ID}, "")
		}(i)
	}

	start.Done()
	done.Wait()

	allowed := 0

	for _, recorder := range recorders {
		assert.Assert(t, is.Equal(http.StatusOK, recorder.Code), recorder.Body.String())

		var out models.ShareLinkConsumeResponse
		assert.NilError(t, json.Unmarshal(recorder.Body.Bytes(), &out))

		switch {
		case out.Allowed:
			allowed++
		default:
			assert.Check(t, is.Equal(models.ShareLinkConsumeReasonExhausted, out.Reason))
		}
	}

	assert.Check(t, is.Equal(1, allowed), "exactly one concurrent download passes a single use share")

	stored, err := suite.db.ShareLink.Get(fx.ctx, link.ID)
	assert.NilError(t, err)
	assert.Check(t, is.Equal(1, stored.UseCount))
	assert.Check(t, is.Len(suite.shareLinkAccessEvents(fx, link.ID), allowed), "one access event row per allowed download")
}
