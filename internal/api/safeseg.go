package api

import (
	"net/url"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// safeSeg validates an ID arg as a URL-safe path segment and returns the
// PathEscape-d form ready for interpolation. Used by every endpoint method
// that interpolates user-supplied IDs into a path; without it, an
// adversarial caller could send `bar_id="../../strategy/objectives/SECRET"`
// and pivot a manage_bar action to a different resource.
//
// PathEscape on a validator-approved ID is a no-op today (the regex restricts
// to URL-safe chars), but it stays as belt-and-braces against future regex
// loosening.
func safeSeg(field, value string) (string, error) {
	if err := productplan.Field(field).RequireID(value); err != nil {
		return "", err
	}
	return url.PathEscape(strings.TrimSpace(value)), nil
}

// safeSegPair is a convenience for the recurring "validate two IDs, then
// interpolate" pattern used by every update/delete of a sub-resource. It
// returns the two escaped segments or the first validation error encountered.
// The order of fields in the call site is the order returned, which keeps the
// URL composition local and obvious at the call site.
func safeSegPair(field1, value1, field2, value2 string) (string, string, error) {
	seg1, err := safeSeg(field1, value1)
	if err != nil {
		return "", "", err
	}
	seg2, err := safeSeg(field2, value2)
	if err != nil {
		return "", "", err
	}
	return seg1, seg2, nil
}
