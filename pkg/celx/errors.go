package celx

import "errors"

// ErrJSONMapExpected indicates a CEL value could not be converted into a JSON object map

var ErrJSONMapExpected = errors.New("expected JSON object")
