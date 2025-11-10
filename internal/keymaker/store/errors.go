package store

import "errors"

var ErrMissingIdentifiers = errors.New("keymaker: missing org or integration id")
