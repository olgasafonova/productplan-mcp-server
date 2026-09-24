package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func argTool() Tool {
	return Tool{Name: "manage_thing", InputSchema: InputSchema{
		Type: "object",
		Properties: map[string]Property{
			"action":  {Type: "string"},
			"legend":  {Type: "string"},
			"lane_id": {Type: "string"},
		},
	}}
}

func TestCheckArgumentKeys(t *testing.T) {
	tool := argTool()
	for name, tc := range map[string]struct {
		args    map[string]any
		wantErr bool
		want    []string
		notWant []string
	}{
		"nil args":        {args: nil},
		"empty args":      {args: map[string]any{}},
		"all known":       {args: map[string]any{"action": "create", "legend": "x", "lane_id": "1"}},
		"typo suggested":  {args: map[string]any{"legnd": "x"}, wantErr: true, want: []string{`unknown argument for manage_thing: "legnd" (did you mean "legend"?)`, "Valid arguments: action, lane_id, legend"}},
		"far key no hint": {args: map[string]any{"zzzzzz": 1}, wantErr: true, want: []string{`"zzzzzz"`}, notWant: []string{"did you mean"}},
		"several sorted":  {args: map[string]any{"b_unknown": 1, "a_unknown": 1, "action": "x"}, wantErr: true, want: []string{`unknown arguments for manage_thing: "a_unknown", "b_unknown"`}},
	} {
		t.Run(name, func(t *testing.T) {
			err := tool.CheckArgumentKeys(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil {
				return
			}
			for _, w := range tc.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error %q lacks %q", err, w)
				}
			}
			for _, w := range tc.notWant {
				if strings.Contains(err.Error(), w) {
					t.Errorf("error %q should not contain %q", err, w)
				}
			}
		})
	}
}

func TestCheckArgumentKeys_NoArgumentTool(t *testing.T) {
	err := Tool{Name: "check_status"}.CheckArgumentKeys(map[string]any{"deep": true})
	if err == nil || !strings.Contains(err.Error(), "check_status takes no arguments") {
		t.Fatalf("got %v", err)
	}
}

func TestEditDistance(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"", "", 0}, {"abc", "", 3}, {"", "ab", 2}, {"legend", "legend", 0},
		{"legnd", "legend", 1}, {"lane", "lnae", 2}, {"kitten", "sitting", 3},
	} {
		if got := editDistance(tc.a, tc.b); got != tc.want {
			t.Errorf("editDistance(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

// The handler must not run when an argument is unknown.
func TestRegistryCall_RejectsUnknownKeyBeforeHandler(t *testing.T) {
	r := NewRegistry()
	ran := false
	r.RegisterFunc(argTool(), func(context.Context, map[string]any) (json.RawMessage, error) {
		ran = true
		return json.RawMessage(`{}`), nil
	})
	if _, err := r.Call(context.Background(), "manage_thing", map[string]any{"colour": "red"}); err == nil {
		t.Fatal("unknown key accepted")
	}
	if ran {
		t.Error("handler ran despite an unknown argument")
	}
	if _, err := r.Call(context.Background(), "manage_thing", map[string]any{"action": "create"}); err != nil {
		t.Fatalf("known key rejected: %v", err)
	}
	if !ran {
		t.Error("handler did not run for known keys")
	}
}
