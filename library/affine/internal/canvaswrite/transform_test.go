package canvaswrite

import "testing"

func TestBuildTransformPlanMoveResizeAndDisplayMode(t *testing.T) {
	selectors := SearchResult{
		DocID:      "doc-1",
		SourceMode: "snapshot",
		Count:      1,
		Entities: []SearchEntity{
			{ID: "card", Kind: "card", DisplayMode: "edgeless", XYWH: []float64{10, 20, 100, 80}},
		},
	}
	plan, err := BuildTransformPlan(selectors, TransformOptions{
		DocID:       "doc-1",
		Move:        "5,-10",
		Resize:      "120,90",
		DisplayMode: "embed",
	})
	if err != nil {
		t.Fatalf("BuildTransformPlan error: %v", err)
	}
	if plan.PlanType != "canvas_transform" || !plan.DryRun {
		t.Fatalf("plan header = %#v, want canvas_transform dry-run", plan)
	}
	if len(plan.Operations) != 3 {
		t.Fatalf("operations = %d, want 3", len(plan.Operations))
	}
	if plan.AffectedIDs[0] != "card" {
		t.Fatalf("affected = %#v, want card", plan.AffectedIDs)
	}
}

func TestBuildTransformPlanDirectIDAllowsDisplayMode(t *testing.T) {
	plan, err := BuildTransformPlan(SearchResult{SourceMode: "direct"}, TransformOptions{
		IDs:         []string{"card"},
		DisplayMode: "embed",
	})
	if err != nil {
		t.Fatalf("BuildTransformPlan error: %v", err)
	}
	if len(plan.Operations) != 1 || plan.Operations[0].ID != "card" {
		t.Fatalf("operations = %#v, want display mode op for direct id", plan.Operations)
	}
}

func TestBuildTransformPlanRejectsGeometryWithoutXYWH(t *testing.T) {
	_, err := BuildTransformPlan(SearchResult{
		SourceMode: "snapshot",
		Entities:   []SearchEntity{{ID: "card", Kind: "card"}},
	}, TransformOptions{Move: "1,1"})
	if err == nil {
		t.Fatal("BuildTransformPlan error = nil, want xywh error")
	}
}

func TestBuildTransformPlanRejectsIntegrityIssues(t *testing.T) {
	_, err := BuildTransformPlan(SearchResult{
		SourceMode: "snapshot",
		Entities:   []SearchEntity{{ID: "card", Kind: "card", IntegrityIssues: []string{"missing_children"}}},
	}, TransformOptions{DisplayMode: "embed"})
	if err == nil {
		t.Fatal("BuildTransformPlan error = nil, want integrity error")
	}
}
