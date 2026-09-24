package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestSafeSeg_RejectsInjection is the regression test for the path-injection
// finding (Carlini scan, productplan Finding 2). Each payload must be
// rejected before the request reaches the API.
func TestSafeSeg_RejectsInjection(t *testing.T) {
	cases := []struct {
		name string
		id   string
	}{
		{"path traversal", "../../strategy/objectives/SECRET"},
		{"sub-resource pivot", "X/comments/Y"},
		{"query injection", "123?expand=*&force=true"},
		{"hash anchor", "valid#anchor"},
		{"semicolon param", "valid;param"},
		{"ampersand", "valid&query=injected"},
		{"newline injection", "valid\nX-Inject: 1"},
		{"null byte", "valid\x00.attacker"},
		{"empty", ""},
		{"only whitespace", "   "},
		{"oversize", strings.Repeat("a", 101)},
		{"forward slash", "a/b"},
		{"percent encoding", "valid%2Fextra"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			seg, err := safeSeg("test_id", c.id)
			if err == nil {
				t.Errorf("safeSeg(%q) returned nil error; expected rejection. got seg=%q", c.id, seg)
			}
		})
	}
}

// TestSafeSeg_AcceptsRealisticIDs guards the success path. ProductPlan IDs
// in the wild are short alphanumeric tokens, sometimes with hyphens or
// underscores. None of these should be rejected by the validator.
func TestSafeSeg_AcceptsRealisticIDs(t *testing.T) {
	realisticIDs := []string{
		"wbihRzEYTdOOOLXTeyc8",
		"abc123",
		"user_id_123",
		"hyphen-id-456",
		"BAR-789",
		"123",
		"a",
	}

	for _, id := range realisticIDs {
		t.Run(id, func(t *testing.T) {
			seg, err := safeSeg("test_id", id)
			if err != nil {
				t.Errorf("safeSeg(%q) returned %v; expected nil for realistic ID", id, err)
			}
			if seg == "" {
				t.Errorf("safeSeg(%q) returned empty string", id)
			}
		})
	}
}

// TestIDTypes_NameTheirField pins the field each typed ID reports, which is
// the argument name agents see in validation errors.
func TestIDTypes_NameTheirField(t *testing.T) {
	const bad = "../x"
	cases := []struct {
		id    pathID
		field string
	}{
		{RoadmapID(bad), "roadmap_id"},
		{LaneID(bad), "lane_id"},
		{MilestoneID(bad), "milestone_id"},
		{BarID(bad), "bar_id"},
		{ConnectionID(bad), "connection_id"},
		{LinkID(bad), "link_id"},
		{ObjectiveID(bad), "objective_id"},
		{KeyResultID(bad), "key_result_id"},
		{IdeaID(bad), "idea_id"},
		{OpportunityID(bad), "opportunity_id"},
		{IdeaFormID(bad), "idea_form_id"},
		{LaunchID(bad), "launch_id"},
		{SectionID(bad), "section_id"},
		{TaskID(bad), "task_id"},
	}
	for _, c := range cases {
		t.Run(c.field, func(t *testing.T) {
			_, err := c.id.segment()
			want := "validation error for '" + c.field + "': contains invalid characters"
			if err == nil || !strings.HasPrefix(err.Error(), want) {
				t.Errorf("segment() error = %v, want prefix %q", err, want)
			}
		})
	}
}

// TestRouteWith_ValidatesEveryID checks that a later invalid ID fails the
// path even when earlier ones are fine, and that valid IDs are substituted
// in order.
func TestRouteWith_ValidatesEveryID(t *testing.T) {
	got, err := route("/launches/%s/tasks/%s").with(LaunchID("7"), TaskID(" 9 "))
	if err != nil || got != "/launches/7/tasks/9" {
		t.Fatalf("with() = %q, %v; want /launches/7/tasks/9", got, err)
	}
	if _, err := route("/launches/%s/tasks/%s").with(LaunchID("7"), TaskID("9/../x")); err == nil ||
		!strings.Contains(err.Error(), "task_id") {
		t.Fatalf("with() error = %v; want task_id rejection", err)
	}
}

// TestEndpoint_RejectsInjectionBeforeHTTP exercises the integrated path
// through one representative endpoint method. Asserts the validator rejects
// the malicious bar_id before any HTTP layer is touched (the Client has no
// transport configured, so any request that reaches the wire would error
// differently).
func TestEndpoint_RejectsInjectionBeforeHTTP(t *testing.T) {
	c := &Client{baseURL: "https://example.invalid"}

	maliciousID := "../../strategy/objectives/SECRET"
	_, err := c.GetBar(context.Background(), BarID(maliciousID))

	if err == nil {
		t.Fatal("GetBar accepted path-injection payload; expected validator rejection")
	}
	if !strings.Contains(err.Error(), "bar_id") {
		t.Errorf("error did not mention bar_id: %v", err)
	}
	if strings.Contains(err.Error(), "SECRET") {
		t.Errorf("error leaked the injected payload: %v", err)
	}
}

// TestEndpoint_RejectsInjectionInSecondID covers the multi-ID case (e.g.,
// DeleteKeyResult takes both objective_id and key_result_id). A malicious
// caller might supply a valid first ID and a malicious second.
func TestEndpoint_RejectsInjectionInSecondID(t *testing.T) {
	c := &Client{baseURL: "https://example.invalid"}

	_, err := c.DeleteKeyResult(context.Background(), "validObjective", "../launches/X")

	if err == nil {
		t.Fatal("DeleteKeyResult accepted injection in key_result_id; expected validator rejection")
	}
	if !strings.Contains(err.Error(), "key_result_id") {
		t.Errorf("error did not mention key_result_id: %v", err)
	}
}

// TestEveryIDEndpoint_RejectsInjectionBeforeHTTP walks every endpoint method
// that takes an ID, feeding "../" in each ID position in turn, against a
// server that counts requests. Every call must fail naming the bad field,
// and the server must see nothing.
func TestEveryIDEndpoint_RejectsInjectionBeforeHTTP(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()
	c := testClient(t, srv)
	ctx := context.Background()
	const bad, ok = "../../strategy/objectives/SECRET", "1"
	body := map[string]any{}

	type call = func() (json.RawMessage, error)
	cases := []struct {
		name  string
		field string
		call  call
	}{
		{"GetRoadmap", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmap(ctx, bad) }},
		{"GetRoadmapBars", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmapBars(ctx, bad) }},
		{"GetRoadmapLanes", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmapLanes(ctx, bad) }},
		{"GetRoadmapMilestones", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmapMilestones(ctx, bad) }},
		{"GetRoadmapLegends", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmapLegends(ctx, bad) }},
		{"GetRoadmapComments", "roadmap_id", func() (json.RawMessage, error) { return c.GetRoadmapComments(ctx, bad) }},
		{"CreateLane", "roadmap_id", func() (json.RawMessage, error) { return c.CreateLane(ctx, bad, body) }},
		{"UpdateLane/roadmap", "roadmap_id", func() (json.RawMessage, error) { return c.UpdateLane(ctx, bad, ok, body) }},
		{"UpdateLane/lane", "lane_id", func() (json.RawMessage, error) { return c.UpdateLane(ctx, ok, bad, body) }},
		{"DeleteLane/roadmap", "roadmap_id", func() (json.RawMessage, error) { return c.DeleteLane(ctx, bad, ok) }},
		{"DeleteLane/lane", "lane_id", func() (json.RawMessage, error) { return c.DeleteLane(ctx, ok, bad) }},
		{"CreateMilestone", "roadmap_id", func() (json.RawMessage, error) { return c.CreateMilestone(ctx, bad, body) }},
		{"UpdateMilestone/roadmap", "roadmap_id", func() (json.RawMessage, error) { return c.UpdateMilestone(ctx, bad, ok, body) }},
		{"UpdateMilestone/milestone", "milestone_id", func() (json.RawMessage, error) { return c.UpdateMilestone(ctx, ok, bad, body) }},
		{"DeleteMilestone/roadmap", "roadmap_id", func() (json.RawMessage, error) { return c.DeleteMilestone(ctx, bad, ok) }},
		{"DeleteMilestone/milestone", "milestone_id", func() (json.RawMessage, error) { return c.DeleteMilestone(ctx, ok, bad) }},
		{"GetBar", "bar_id", func() (json.RawMessage, error) { return c.GetBar(ctx, bad) }},
		{"UpdateBar", "bar_id", func() (json.RawMessage, error) { return c.UpdateBar(ctx, bad, body) }},
		{"DeleteBar", "bar_id", func() (json.RawMessage, error) { return c.DeleteBar(ctx, bad) }},
		{"GetBarChildren", "bar_id", func() (json.RawMessage, error) { return c.GetBarChildren(ctx, bad) }},
		{"GetBarComments", "bar_id", func() (json.RawMessage, error) { return c.GetBarComments(ctx, bad) }},
		{"GetBarConnections", "bar_id", func() (json.RawMessage, error) { return c.GetBarConnections(ctx, bad) }},
		{"CreateBarConnection", "bar_id", func() (json.RawMessage, error) { return c.CreateBarConnection(ctx, bad, body) }},
		{"DeleteBarConnection/bar", "bar_id", func() (json.RawMessage, error) { return c.DeleteBarConnection(ctx, bad, ok) }},
		{"DeleteBarConnection/connection", "connection_id", func() (json.RawMessage, error) { return c.DeleteBarConnection(ctx, ok, bad) }},
		{"GetBarLinks", "bar_id", func() (json.RawMessage, error) { return c.GetBarLinks(ctx, bad) }},
		{"CreateBarLink", "bar_id", func() (json.RawMessage, error) { return c.CreateBarLink(ctx, bad, body) }},
		{"DeleteBarLink/bar", "bar_id", func() (json.RawMessage, error) { return c.DeleteBarLink(ctx, bad, ok) }},
		{"DeleteBarLink/link", "link_id", func() (json.RawMessage, error) { return c.DeleteBarLink(ctx, ok, bad) }},
		{"GetObjective", "objective_id", func() (json.RawMessage, error) { return c.GetObjective(ctx, bad) }},
		{"UpdateObjective", "objective_id", func() (json.RawMessage, error) { return c.UpdateObjective(ctx, bad, body) }},
		{"DeleteObjective", "objective_id", func() (json.RawMessage, error) { return c.DeleteObjective(ctx, bad) }},
		{"ListKeyResults", "objective_id", func() (json.RawMessage, error) { return c.ListKeyResults(ctx, bad) }},
		{"GetKeyResult/objective", "objective_id", func() (json.RawMessage, error) { return c.GetKeyResult(ctx, bad, ok) }},
		{"GetKeyResult/key_result", "key_result_id", func() (json.RawMessage, error) { return c.GetKeyResult(ctx, ok, bad) }},
		{"CreateKeyResult", "objective_id", func() (json.RawMessage, error) { return c.CreateKeyResult(ctx, bad, body) }},
		{"UpdateKeyResult/objective", "objective_id", func() (json.RawMessage, error) { return c.UpdateKeyResult(ctx, bad, ok, body) }},
		{"UpdateKeyResult/key_result", "key_result_id", func() (json.RawMessage, error) { return c.UpdateKeyResult(ctx, ok, bad, body) }},
		{"DeleteKeyResult/objective", "objective_id", func() (json.RawMessage, error) { return c.DeleteKeyResult(ctx, bad, ok) }},
		{"DeleteKeyResult/key_result", "key_result_id", func() (json.RawMessage, error) { return c.DeleteKeyResult(ctx, ok, bad) }},
		{"GetIdea", "idea_id", func() (json.RawMessage, error) { return c.GetIdea(ctx, bad) }},
		{"UpdateIdea", "idea_id", func() (json.RawMessage, error) { return c.UpdateIdea(ctx, bad, body) }},
		{"GetOpportunity", "opportunity_id", func() (json.RawMessage, error) { return c.GetOpportunity(ctx, bad) }},
		{"UpdateOpportunity", "opportunity_id", func() (json.RawMessage, error) { return c.UpdateOpportunity(ctx, bad, body) }},
		{"GetIdeaForm", "idea_form_id", func() (json.RawMessage, error) { return c.GetIdeaForm(ctx, bad) }},
		{"GetLaunch", "launch_id", func() (json.RawMessage, error) { return c.GetLaunch(ctx, bad) }},
		{"UpdateLaunch", "launch_id", func() (json.RawMessage, error) { return c.UpdateLaunch(ctx, bad, body) }},
		{"DeleteLaunch", "launch_id", func() (json.RawMessage, error) { return c.DeleteLaunch(ctx, bad) }},
		{"GetLaunchSections", "launch_id", func() (json.RawMessage, error) { return c.GetLaunchSections(ctx, bad) }},
		{"GetLaunchSection/launch", "launch_id", func() (json.RawMessage, error) { return c.GetLaunchSection(ctx, bad, ok) }},
		{"GetLaunchSection/section", "section_id", func() (json.RawMessage, error) { return c.GetLaunchSection(ctx, ok, bad) }},
		{"CreateLaunchSection", "launch_id", func() (json.RawMessage, error) { return c.CreateLaunchSection(ctx, bad, body) }},
		{"UpdateLaunchSection/launch", "launch_id", func() (json.RawMessage, error) { return c.UpdateLaunchSection(ctx, bad, ok, body) }},
		{"UpdateLaunchSection/section", "section_id", func() (json.RawMessage, error) { return c.UpdateLaunchSection(ctx, ok, bad, body) }},
		{"DeleteLaunchSection/launch", "launch_id", func() (json.RawMessage, error) { return c.DeleteLaunchSection(ctx, bad, ok) }},
		{"DeleteLaunchSection/section", "section_id", func() (json.RawMessage, error) { return c.DeleteLaunchSection(ctx, ok, bad) }},
		{"GetLaunchTasks", "launch_id", func() (json.RawMessage, error) { return c.GetLaunchTasks(ctx, bad) }},
		{"GetLaunchTask/launch", "launch_id", func() (json.RawMessage, error) { return c.GetLaunchTask(ctx, bad, ok) }},
		{"GetLaunchTask/task", "task_id", func() (json.RawMessage, error) { return c.GetLaunchTask(ctx, ok, bad) }},
		{"CreateLaunchTask", "launch_id", func() (json.RawMessage, error) { return c.CreateLaunchTask(ctx, bad, body) }},
		{"UpdateLaunchTask/launch", "launch_id", func() (json.RawMessage, error) { return c.UpdateLaunchTask(ctx, bad, ok, body) }},
		{"UpdateLaunchTask/task", "task_id", func() (json.RawMessage, error) { return c.UpdateLaunchTask(ctx, ok, bad, body) }},
		{"DeleteLaunchTask/launch", "launch_id", func() (json.RawMessage, error) { return c.DeleteLaunchTask(ctx, bad, ok) }},
		{"DeleteLaunchTask/task", "task_id", func() (json.RawMessage, error) { return c.DeleteLaunchTask(ctx, ok, bad) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.call()
			if err == nil || !strings.Contains(err.Error(), "'"+tc.field+"'") {
				t.Errorf("error = %v, want a %s validation error", err, tc.field)
			}
		})
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server received %d requests; every injected ID must fail before HTTP", n)
	}
}
