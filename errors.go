package zip2jsons

import "errors"

// ErrNewReader indicates a failure to create a new zip reader, often due to invalid
// or truncated zip data.
var ErrNewReader = errors.New("failed to create new zip reader")
