package api

import (
	"net/url"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// safeSeg validates an ID as a URL-safe path segment and returns the
// PathEscape-d form ready for interpolation. It is the one place a typed
// ID (ids.go) becomes path text: route.with only accepts pathID values, and
// every pathID's segment method calls safeSeg with its own field name.
// Without it, an adversarial caller could send
// `bar_id="../../strategy/objectives/SECRET"` and pivot a manage_bar action
// to a different resource.
//
// PathEscape on a validator-approved ID is a no-op today (the regex restricts
// to URL-safe chars), but it stays as belt-and-braces against future regex
// loosening.
func safeSeg(field productplan.Field, value string) (string, error) {
	if err := field.RequireID(value); err != nil {
		return "", err
	}
	return url.PathEscape(strings.TrimSpace(value)), nil
}
