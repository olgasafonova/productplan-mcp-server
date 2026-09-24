package tools

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/olgasafonova/productplan-mcp-server/internal/api"
)

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestBarFieldsPayload(t *testing.T) {
	tests := []struct {
		name   string
		fields BarFields
		want   map[string]any
	}{
		{"omitted fields are never sent", BarFields{}, map[string]any{}},
		{"every documented field", BarFields{
			Name: "N", Description: "D", StartsOn: "2026-01-01", EndsOn: "2026-02-01",
			StrategicValue: "S", Notes: "No", PercentDone: intPtr(0), Tags: []string{"a"},
			Lane: "Backend", Legend: strPtr("Committed"), IsContainer: boolPtr(true),
			Parked: boolPtr(false), ContainerBarID: "42",
			CustomTextFields:     []CustomFieldValue{{Label: "Owner", Value: "Olga"}},
			CustomDropdownFields: []CustomFieldValue{{Label: "Objectives", Value: "Grow"}},
		}, map[string]any{
			"name": "N", "description": "D", "starts_on": "2026-01-01", "ends_on": "2026-02-01",
			"strategic_value": "S", "notes": "No", "percent_done": 0, "tags": []string{"a"},
			"lane": "Backend", "legend": "Committed", "is_container": true, "parked": false,
			"container_bar_id":       int64(42),
			"custom_text_fields":     []map[string]any{{"label": "Owner", "value": "Olga"}},
			"custom_dropdown_fields": []map[string]any{{"label": "Objectives", "value": "Grow"}},
		}},
		{"legend empty string clears to null", BarFields{Legend: strPtr("")}, map[string]any{"legend": nil}},
		{"clear_legend clears to null", BarFields{ClearLegend: true}, map[string]any{"legend": nil}},
		{"explicit empty tags clears the list", BarFields{Tags: []string{}}, map[string]any{"tags": []string{}}},
		{"lane_id sent as integer", BarFields{LaneID: "77"}, map[string]any{"lane_id": int64(77)}},
		{"parent_id maps to container_bar_id", BarFields{ParentID: "9"}, map[string]any{"container_bar_id": int64(9)}},
		{"container maps to is_container", BarFields{Container: boolPtr(false)}, map[string]any{"is_container": false}},
		{"custom field name maps to label", BarFields{CustomTextFields: []CustomFieldValue{{Name: "Owner", Value: "x"}}},
			map[string]any{"custom_text_fields": []map[string]any{{"label": "Owner", "value": "x"}}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fields.payload()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("payload mismatch\n got: %#v\nwant: %#v", got, tt.want)
			}
			for _, banned := range []string{"legend_id", "parent_id", "container", "effort"} {
				if _, ok := got[banned]; ok {
					t.Errorf("payload forwarded %q, which the API ignores or misreads", banned)
				}
			}
		})
	}
}

func TestBarFieldsPayloadRejects(t *testing.T) {
	tests := []struct {
		name    string
		fields  BarFields
		wantErr string
	}{
		{"legend_id", BarFields{LegendID: "1"}, "CLEAR the bar's color"},
		{"effort", BarFields{Effort: intPtr(3)}, "effort is not a ProductPlan bar field"},
		{"lane and lane_id", BarFields{Lane: "Backend", LaneID: "1"}, "not both"},
		{"bad date", BarFields{StartsOn: "01-02-2026"}, "YYYY-MM-DD"},
		{"reversed dates", BarFields{StartsOn: "2026-03-01", EndsOn: "2026-02-01"}, "after ends_on"},
		{"percent out of range", BarFields{PercentDone: intPtr(101)}, "0-100"},
		{"clear_legend vs legend", BarFields{ClearLegend: true, Legend: strPtr("Committed")}, "contradicts"},
		{"alias disagreement", BarFields{ParentID: "1", ContainerBarID: "2"}, "disagree"},
		{"non-numeric container", BarFields{ContainerBarID: "abc"}, "numeric"},
		{"custom field without label", BarFields{CustomTextFields: []CustomFieldValue{{Value: "x"}}}, "needs a label"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.fields.payload()
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("want error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestResolveAgainstSchema(t *testing.T) {
	schema := &api.BarWriteSchema{
		RoadmapID: "100", Legends: []string{"Committed"}, Lanes: []string{"Backend"},
		CustomTextFields:     []string{"Owner"},
		CustomDropdownFields: []api.DropdownField{{Label: "Objectives", AllowedValues: []string{"Grow"}}},
	}
	t.Run("canonicalises case-insensitive matches", func(t *testing.T) {
		p := map[string]any{
			"legend": "committed", "lane": "BACKEND",
			"custom_text_fields":     []map[string]any{{"label": "owner", "value": "x"}},
			"custom_dropdown_fields": []map[string]any{{"label": "objectives", "value": "grow"}},
		}
		if err := resolveAgainstSchema(p, schema); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p["legend"] != "Committed" || p["lane"] != "Backend" {
			t.Errorf("names not canonicalised: %v", p)
		}
		dd := p["custom_dropdown_fields"].([]map[string]any)[0]
		if dd["label"] != "Objectives" || dd["value"] != "Grow" {
			t.Errorf("dropdown not canonicalised: %v", dd)
		}
	})
	t.Run("reports every problem with the valid options", func(t *testing.T) {
		p := map[string]any{
			"legend": "Purple", "lane": "Web",
			"custom_text_fields":     []map[string]any{{"label": "Nope", "value": "x"}},
			"custom_dropdown_fields": []map[string]any{{"label": "Objectives", "value": "Shrink"}},
		}
		err := resolveAgainstSchema(p, schema)
		if err == nil {
			t.Fatal("expected error")
		}
		for _, want := range []string{`legend "Purple"`, `Valid legends: "Committed"`, `lane "Web"`, `Valid lanes: "Backend"`,
			`custom text field "Nope"`, `"Shrink" is not an allowed value`, `Allowed values: "Grow"`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error missing %q: %v", want, err)
			}
		}
	})
	t.Run("null legend (clear) needs no lookup", func(t *testing.T) {
		if needsSchema(map[string]any{"legend": nil}) {
			t.Error("clearing the legend should not require a roadmap fetch")
		}
	})
}

func callManageBar(t *testing.T, client *api.Client, args map[string]any) (FormattedResponse, error) {
	t.Helper()
	out, err := manageBarHandler(client).Handle(context.Background(), args)
	var fr FormattedResponse
	if err == nil {
		if uerr := json.Unmarshal(out, &fr); uerr != nil {
			t.Fatalf("bad response: %v", uerr)
		}
	}
	return fr, err
}

func TestManageBarCreate(t *testing.T) {
	f := newFakeAPI()
	client := f.start(t)
	fr, err := callManageBar(t, client, map[string]any{
		"action": "create", "roadmap_id": "100", "name": "New", "lane": "backend",
		"legend": "Committed", "starts_on": "2026-01-01", "ends_on": "2026-02-01",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(fr.Data, &data); err != nil {
		t.Fatal(err)
	}
	if data["id"] != "5001" {
		t.Errorf("expected id parsed from location, got %v", data["id"])
	}
	if data["bar"] == nil {
		t.Error("expected bar read back after create")
	}
	w := f.writes()
	if len(w) != 1 || w[0].Method != "POST" {
		t.Fatalf("expected one POST, got %+v", w)
	}
	body := w[0].Body
	if body["lane"] != "Backend" || body["legend"] != "Committed" || body["parked"] != false || body["roadmap_id"] != float64(100) {
		t.Errorf("unexpected create body: %v", body)
	}
	if _, ok := body["lane_id"]; ok {
		t.Error("lane_id sent although caller gave a lane name")
	}
}

func TestManageBarCreateUndatedStaysParkedByDefault(t *testing.T) {
	f := newFakeAPI()
	client := f.start(t)
	fr, err := callManageBar(t, client, map[string]any{"action": "create", "roadmap_id": "100", "name": "N", "lane_id": "7"})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.writes()[0].Body["parked"]; ok {
		t.Error("parked must not be sent when the caller gave no dates and no parked")
	}
	if !strings.Contains(fr.Summary, "parked") {
		t.Errorf("summary should warn the bar is parked: %s", fr.Summary)
	}
	if f.countGets("/roadmaps/100") != 0 {
		t.Error("roadmap fetched although nothing needed validating")
	}
}

func TestManageBarCreateNestedInheritsParked(t *testing.T) {
	f := newFakeAPI()
	f.addBar("42", map[string]any{"parked": true, "starts_on": "2026-01-01", "ends_on": "2026-03-01"})
	client := f.start(t)
	_, err := callManageBar(t, client, map[string]any{"action": "create", "roadmap_id": "100", "name": "Child",
		"lane_id": "7", "container_bar_id": "42", "starts_on": "2026-01-01", "ends_on": "2026-02-01"})
	if err != nil {
		t.Fatal(err)
	}
	if got := f.writes()[0].Body["parked"]; got != true {
		t.Errorf("nested bar should inherit container parked=true, got %v", got)
	}

	_, err = callManageBar(t, client, map[string]any{"action": "create", "roadmap_id": "100", "name": "Undated",
		"lane_id": "7", "container_bar_id": "42"})
	if err == nil || !strings.Contains(err.Error(), "requires starts_on and ends_on") {
		t.Errorf("expected date pre-check error, got %v", err)
	}
}

func TestManageBarUpdateLegendIDRejected(t *testing.T) {
	f := newFakeAPI()
	f.addBar("789", nil)
	client := f.start(t)
	_, err := callManageBar(t, client, map[string]any{"action": "update", "bar_id": "789", "legend_id": "1"})
	if err == nil {
		t.Fatal("legend_id must be rejected")
	}
	for _, want := range []string{"CLEAR the bar's color", "`legend`", `"Committed", "Exploring", "Blocked"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error missing %q: %v", want, err)
		}
	}
	if len(f.writes()) != 0 {
		t.Error("nothing may be written when legend_id is rejected")
	}
}

func TestManageBarUpdate(t *testing.T) {
	f := newFakeAPI()
	f.addBar("789", map[string]any{"parked": false})
	f.addBar("42", map[string]any{"parked": true})
	client := f.start(t)

	t.Run("unknown legend fails before writing", func(t *testing.T) {
		_, err := callManageBar(t, client, map[string]any{"action": "update", "bar_id": "789", "legend": "Purple"})
		if err == nil || !strings.Contains(err.Error(), "Valid legends") {
			t.Errorf("expected legend validation error, got %v", err)
		}
	})
	t.Run("clear legend sends null", func(t *testing.T) {
		if _, err := callManageBar(t, client, map[string]any{"action": "update", "bar_id": "789", "legend": ""}); err != nil {
			t.Fatal(err)
		}
		w := f.writes()
		body := w[len(w)-1].Body
		if v, ok := body["legend"]; !ok || v != nil || len(body) != 1 {
			t.Errorf("expected exactly {legend:null}, got %v", body)
		}
	})
	t.Run("parked mismatch with container fails early", func(t *testing.T) {
		_, err := callManageBar(t, client, map[string]any{"action": "update", "bar_id": "789",
			"container_bar_id": "42", "starts_on": "2026-01-01", "ends_on": "2026-02-01"})
		if err == nil || !strings.Contains(err.Error(), "pass parked:true") {
			t.Errorf("expected parked pre-check, got %v", err)
		}
	})
	t.Run("empty update refused", func(t *testing.T) {
		_, err := callManageBar(t, client, map[string]any{"action": "update", "bar_id": "789"})
		if err == nil || !strings.Contains(err.Error(), "nothing to update") {
			t.Errorf("expected nothing-to-update error, got %v", err)
		}
	})
	if n := len(f.writes()); n != 1 {
		t.Errorf("expected only the legend-clear PATCH to be sent, got %d writes", n)
	}
}

func TestParseCreatedID(t *testing.T) {
	tests := []struct {
		body, want string
		wantErr    bool
	}{
		{`{"location":"/api/v2/bars/36935827"}`, "36935827", false},
		{`{"location":"/api/v2/bars/12/"}`, "12", false},
		{`{"id":55}`, "55", false},
		{`{"location":"/api/v2/bars/../x?y"}`, "", true},
		{`{}`, "", true},
	}
	for _, tt := range tests {
		got, err := api.ParseCreatedID(json.RawMessage(tt.body))
		if (err != nil) != tt.wantErr || got != tt.want {
			t.Errorf("ParseCreatedID(%s) = %q, %v; want %q, err=%v", tt.body, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestGetRoadmapLegendsReturnsNames(t *testing.T) {
	f := newFakeAPI()
	client := f.start(t)
	out, err := getRoadmapLegendsHandler(client).Handle(context.Background(), map[string]any{"roadmap_id": "100"})
	if err != nil {
		t.Fatal(err)
	}
	var fr FormattedResponse
	if err := json.Unmarshal(out, &fr); err != nil {
		t.Fatal(err)
	}
	var names []string
	if err := json.Unmarshal(fr.Data, &names); err != nil {
		t.Fatalf("legends should be a list of names: %v (%s)", err, fr.Data)
	}
	if !reflect.DeepEqual(names, []string{"Committed", "Exploring", "Blocked"}) {
		t.Errorf("unexpected legends %v", names)
	}
	if !strings.Contains(fr.Summary, "Found 3 legends") || !strings.Contains(fr.Summary, "`legend`") {
		t.Errorf("summary should count and explain use: %s", fr.Summary)
	}
}
