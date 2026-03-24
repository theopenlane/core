package types

import "github.com/samber/lo"

// ConnectionRegistration describes one connection mode for a definition
type ConnectionRegistration struct {
	// CredentialRef is the user-facing credential schema that selects this connection mode
	CredentialRef CredentialSlotID `json:"credentialRef"`
	// Name is the user-facing connection mode name
	Name string `json:"name,omitempty"`
	// Description explains what the connection mode does
	Description string `json:"description,omitempty"`
	// CredentialRefs lists the credential slots used by this connection mode
	CredentialRefs []CredentialSlotID `json:"credentialRefs,omitempty"`
	// ClientRefs lists the clients initialized by this connection mode
	ClientRefs []ClientID `json:"-"`
	// ValidationOperation names the operation used to validate credentials before persistence
	ValidationOperation string `json:"validationOperation,omitempty"`
	// Installation describes installation-scoped metadata derived by this connection mode
	Installation *InstallationRegistration `json:"installation,omitempty"`
	// Auth describes how this connection mode performs auth when supported
	Auth *AuthRegistration `json:"auth,omitempty"`
	// Disconnect describes how this connection mode tears down an installation
	Disconnect *DisconnectRegistration `json:"disconnect,omitempty"`
}

// ConnectionRegistration returns the connection registration for the given credential slot
func (d Definition) ConnectionRegistration(ref CredentialSlotID) (ConnectionRegistration, error) {
	reg, found := lo.Find(d.Connections, func(r ConnectionRegistration) bool {
		return r.CredentialRef == ref
	})
	if !found {
		return ConnectionRegistration{}, ErrConnectionRefNotFound
	}

	return reg, nil
}
