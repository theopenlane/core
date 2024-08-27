package sessions

import "errors"

var (
	// ErrInvalidSession is returned when the session is invalid
	ErrInvalidSession = errors.New("invalid session provided")
)
