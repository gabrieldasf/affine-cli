package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestRootIncludesCanvasCommands(t *testing.T) {
	root := RootCmd()
	for _, args := range [][]string{
		{"canvas"},
		{"canvas", "plan"},
		{"canvas", "search"},
		{"canvas", "diff"},
		{"canvas", "transform"},
		{"canvas", "model"},
		{"canvas", "validate"},
		{"canvas", "apply"},
		{"canvas", "example"},
		{"canvas", "doc", "integrity"},
		{"canvas", "doc", "repair"},
	} {
		cmd, _, err := root.Find(args)
		if err != nil {
			t.Fatalf("Find(%v) error: %v", args, err)
		}
		if cmd == nil || cmd.Name() != args[len(args)-1] {
			t.Fatalf("Find(%v) = %v, want %q", args, cmd, args[len(args)-1])
		}
	}
}

func TestCanvasApplyAcceptsTransformPlan(t *testing.T) {
	root := RootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(`{
		"plan_type": "canvas_transform",
		"plan_id": "canvas-transform-test",
		"dry_run": true,
		"affected_ids": ["card"],
		"operations": [{"kind": "set_display_mode", "id": "card", "after": "embed"}]
	}`))
	root.SetArgs([]string{"canvas", "apply", "--dry-run", "--json"})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute error: %v\n%s", err, out.String())
	}
	var got map[string]any
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if got["plan_type"] != "canvas_transform" || got["live_write_supported"] != false {
		t.Fatalf("apply output = %#v, want transform dry-run no live write", got)
	}
}

func TestBuildCanvasPlanVerticalDefault(t *testing.T) {
	plan := buildCanvasPlan(canvasBuildSpec{
		Frame: "TRABALHO",
		Hub:   "Deploy",
		Levels: [][]string{
			{"Deploy"},
			{"Hetzner", "Contabo"},
		},
	})

	if plan.Orientation != "vertical" {
		t.Fatalf("orientation = %q, want vertical", plan.Orientation)
	}
	if len(plan.Nodes) != 3 {
		t.Fatalf("nodes = %d, want 3", len(plan.Nodes))
	}
	if plan.Nodes[0].Role != "hub" {
		t.Fatalf("hub role = %q", plan.Nodes[0].Role)
	}
	if !(plan.Nodes[1].Y > plan.Nodes[0].Y) {
		t.Fatalf("vertical child Y = %v, hub Y = %v", plan.Nodes[1].Y, plan.Nodes[0].Y)
	}
	if len(plan.Connections) != 2 {
		t.Fatalf("connections = %d, want 2", len(plan.Connections))
	}
}

func TestBuildCanvasPlanHorizontal(t *testing.T) {
	plan := buildCanvasPlan(canvasBuildSpec{
		Orientation: "horizontal",
		Hub:         "Coolify",
		Levels: [][]string{
			{"Coolify"},
			{"App A", "App B"},
		},
	})

	if plan.Orientation != "horizontal" {
		t.Fatalf("orientation = %q, want horizontal", plan.Orientation)
	}
	if !(plan.Nodes[1].X > plan.Nodes[0].X) {
		t.Fatalf("horizontal child X = %v, hub X = %v", plan.Nodes[1].X, plan.Nodes[0].X)
	}
}

func TestReadCanvasModelNestedPropsConnection(t *testing.T) {
	input := `{
		"blocks": [
			{"id": "a", "text": "Alpha", "props": {"xywh": "[0,0,360,220]"}},
			{"id": "b", "title": "Beta", "prop:xywh": "[0,360,360,220]"}
		],
		"edges": [
			{"id": "e1", "type": "connector", "props": {"source": "a", "target": "b", "xywh": "[0,0,1,1]"}}
		]
	}`
	model, err := readCanvasModel("", strings.NewReader(input))
	if err != nil {
		t.Fatalf("readCanvasModel error: %v", err)
	}
	if len(model.Nodes) != 2 {
		t.Fatalf("nodes = %d, want 2", len(model.Nodes))
	}
	if len(model.Connections) != 1 {
		t.Fatalf("connections = %d, want 1", len(model.Connections))
	}
	if model.Connections[0]["source"] != "a" || model.Connections[0]["target"] != "b" {
		t.Fatalf("connection = %#v, want source a target b", model.Connections[0])
	}
}

func TestValidateCanvasPlanDetectsOrphanConnection(t *testing.T) {
	report := validateCanvasPlan(canvasPlan{
		Frame:       "TRABALHO",
		Orientation: "vertical",
		Nodes: []canvasNode{
			{ID: "hub", Text: "Hub", W: 360, H: 220},
		},
		Connections: []canvasConnection{
			{From: "Hub", To: "Missing"},
		},
	})

	if report.OK {
		t.Fatalf("OK = true, want false")
	}
	if report.IssueCount == 0 {
		t.Fatalf("IssueCount = 0, want at least 1")
	}
	if report.Issues[0].Code != "orphan_connection" {
		t.Fatalf("first issue code = %q, want orphan_connection", report.Issues[0].Code)
	}
}
