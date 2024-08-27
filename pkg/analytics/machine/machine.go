package machine

import (
	"github.com/panta/machineid"
)

var (
	// ApplicationID is the unique Application ID
	ApplicationID = ""
)

var (
	id string
)

func init() {
	id, _ = machineid.ProtectedID(ApplicationID)
}

func ID() string {
	return id
}

func Available() bool {
	return id != ""
}
