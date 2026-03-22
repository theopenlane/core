package types

// ConnectionRegistration describes one connection mode for a definition
type ConnectionRegistration struct {
	// CredentialRef is the user-facing credential schema that selects this connection mode
	CredentialRef CredentialRef `json:"credentialRef"`
	// Name is the user-facing connection mode name
	Name string `json:"name,omitempty"`
	// Description explains what the connection mode does
	Description string `json:"description,omitempty"`
	// CredentialRefs lists the credential slots used by this connection mode
	CredentialRefs []CredentialRef `json:"credentialRefs,omitempty"`
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

// ConnectionRegistration returns the connection registration for the given credential ref
func (d Definition) ConnectionRegistration(ref CredentialRef) (ConnectionRegistration, error) {
	for _, reg := range d.Connections {
		if reg.CredentialRef.String() == ref.String() {
			return reg, nil
		}
	}

	return ConnectionRegistration{}, ErrConnectionRefNotFound
}
