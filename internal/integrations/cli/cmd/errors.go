package cmd

import "errors"

// ErrUnsupportedOutputFormat is returned when an output format is not supported
var ErrUnsupportedOutputFormat = errors.New("unsupported output format")
