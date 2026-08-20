package serveropts

import "errors"

// ErrNoSigningKeys is returned when no signing keys are found in the key directory
var ErrNoSigningKeys = errors.New("no signing keys found in key directory")

// ErrBackfillRequiresGala is returned when startup backfills are enabled without the gala runtime
var ErrBackfillRequiresGala = errors.New("backfill enabled without gala runtime: enable workflows.gala or disable backfill")
