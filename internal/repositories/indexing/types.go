package indexing

import "errors"

var ErrTableNotFound = errors.New("table not found")
var ErrConditionCheckFailed = errors.New("condition check failed")
var ErrModifyingStaleItem = errors.New("stale item modification attempted")
