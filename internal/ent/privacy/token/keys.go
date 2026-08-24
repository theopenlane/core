package token

import "github.com/theopenlane/utils/contextx"

var (
	signUpTokenContextKey      = contextx.NewKey[*SignUpToken]()
	downloadTokenContextKey    = contextx.NewKey[*DownloadToken]()
	oauthTooTokenContextKey    = contextx.NewKey[*OauthTooToken]()
	orgInviteTokenContextKey   = contextx.NewKey[*OrgInviteToken]()
	resetTokenContextKey       = contextx.NewKey[*ResetToken]()
	verifyTokenContextKey      = contextx.NewKey[*VerifyToken]()
	webauthnCreationContextKey = contextx.NewKey[*WebauthnCreationContextKey]()
)
