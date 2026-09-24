package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/olgasafonova/productplan-mcp-server/pkg/productplan"
)

// DropdownField is one custom dropdown field definition on a roadmap.
type DropdownField struct {
	Label         string   `json:"label"`
	AllowedValues []string `json:"allowed_values"`
}

// BarWriteSchema is the vocabulary a bar write on one roadmap must use:
// legend and lane names, custom text field labels, and custom dropdown
// fields with their allowed values. All of it is embedded in
// GET /roadmaps/{id}; there are no separate endpoints for legends or
// custom field definitions (probed live 24-09-2026).
type BarWriteSchema struct {
	RoadmapID            string          `json:"roadmap_id"`
	Legends              []string        `json:"legends"`
	Lanes                []string        `json:"lanes"`
	CustomTextFields     []string        `json:"custom_text_fields"`
	CustomDropdownFields []DropdownField `json:"custom_dropdown_fields"`
}

// roadmapWriteBody mirrors the parts of the roadmap response that
// BarWriteSchema needs. Legends and lanes arrive as bare name strings;
// labelled entries are decoded leniently so an object shape would not
// break validation.
type roadmapWriteBody struct {
	Legends              []json.RawMessage `json:"legends"`
	Lanes                []json.RawMessage `json:"lanes"`
	CustomTextFields     []json.RawMessage `json:"custom_text_fields"`
	CustomDropdownFields []DropdownField   `json:"custom_dropdown_fields"`
}

// GetBarWriteSchema fetches a roadmap and extracts the names and field
// definitions that bar writes are validated against. It goes through
// GetRoadmap, so any read caching on the client applies automatically.
func (c *Client) GetBarWriteSchema(ctx context.Context, roadmapID RoadmapID) (*BarWriteSchema, error) {
	data, err := c.GetRoadmap(ctx, roadmapID)
	if err != nil {
		return nil, err
	}
	return ParseBarWriteSchema(string(roadmapID), data)
}

// ParseBarWriteSchema extracts a BarWriteSchema from a roadmap payload.
func ParseBarWriteSchema(roadmapID string, data json.RawMessage) (*BarWriteSchema, error) {
	var body roadmapWriteBody
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse roadmap %s response: %w", roadmapID, err)
	}
	return &BarWriteSchema{
		RoadmapID:            roadmapID,
		Legends:              nameList(body.Legends, "name", "label"),
		Lanes:                nameList(body.Lanes, "name"),
		CustomTextFields:     nameList(body.CustomTextFields, "label", "name"),
		CustomDropdownFields: body.CustomDropdownFields,
	}, nil
}

// nameList decodes each entry as a bare string, falling back to the first
// non-empty string among keys when the entry is an object.
func nameList(entries []json.RawMessage, keys ...string) []string {
	names := make([]string, 0, len(entries))
	for _, raw := range entries {
		if name, ok := entryName(raw, keys); ok {
			names = append(names, name)
		}
	}
	return names
}

// entryName reads one entry as a bare string or as an object's first
// non-empty string among keys.
func entryName(raw json.RawMessage, keys []string) (string, bool) {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s, true
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "", false
	}
	for _, k := range keys {
		if v, ok := obj[k].(string); ok && v != "" {
			return v, true
		}
	}
	return "", false
}

// BarSummary is the subset of a bar that write pre-checks need.
type BarSummary struct {
	ID        string
	Name      string
	RoadmapID string
	StartsOn  string
	EndsOn    string
	Parked    *bool
}

// GetBarSummary fetches a bar and extracts the fields write pre-checks use:
// its roadmap, its dates, and whether it is parked.
func (c *Client) GetBarSummary(ctx context.Context, barID BarID) (*BarSummary, error) {
	data, err := c.GetBar(ctx, barID)
	if err != nil {
		return nil, err
	}
	var body struct {
		ID        any    `json:"id"`
		Name      string `json:"name"`
		RoadmapID any    `json:"roadmap_id"`
		Roadmap   struct {
			ID any `json:"id"`
		} `json:"roadmap"`
		StartsOn string `json:"starts_on"`
		EndsOn   string `json:"ends_on"`
		Parked   *bool  `json:"parked"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return nil, fmt.Errorf("failed to parse bar %s response: %w", barID, err)
	}
	roadmapID := idString(body.Roadmap.ID)
	if roadmapID == "" {
		roadmapID = idString(body.RoadmapID)
	}
	return &BarSummary{
		ID:        string(barID),
		Name:      body.Name,
		RoadmapID: roadmapID,
		StartsOn:  body.StartsOn,
		EndsOn:    body.EndsOn,
		Parked:    body.Parked,
	}, nil
}

// idString renders a JSON-decoded ID (number or string) as a string.
func idString(v any) string {
	switch id := v.(type) {
	case string:
		return id
	case float64:
		return strconv.FormatInt(int64(id), 10)
	}
	return ""
}

// ParseCreatedID extracts the new resource's ID from a create response.
// ProductPlan answers POST /bars with {"location":"/api/v2/bars/<id>"}
// rather than the bar itself; an "id" field is accepted as a fallback.
func ParseCreatedID(data json.RawMessage) (string, error) {
	var body struct {
		Location string `json:"location"`
		ID       any    `json:"id"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		return "", fmt.Errorf("failed to parse create response: %w", err)
	}
	if loc := strings.TrimRight(body.Location, "/"); loc != "" {
		id := loc[strings.LastIndex(loc, "/")+1:]
		if err := productplan.Field("id").RequireID(id); err == nil {
			return id, nil
		}
		return "", fmt.Errorf("create response location %q does not end in a valid ID", body.Location)
	}
	if id := idString(body.ID); id != "" {
		return id, nil
	}
	return "", fmt.Errorf("create response carried neither location nor id")
}

// IsNotFound reports whether err is, or wraps, a ProductPlan 404.
// handleResponse wraps *productplan.APIError with %w, so the typed error
// survives the appended suggestion and no message text is parsed.
func IsNotFound(err error) bool {
	var apiErr *productplan.APIError
	return errors.As(err, &apiErr) && apiErr.IsNotFound()
}
