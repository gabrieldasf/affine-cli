package canvaswrite

import "testing"

func TestCheckBlocksIntegrityDetectsMissingChildren(t *testing.T) {
	result := CheckBlocksIntegrity("doc-1", map[string]map[string]any{
		"page": {
			"sys:id":       "page",
			"sys:flavour":  "affine:page",
			"sys:children": []any{"p1"},
		},
		"p1": {
			"sys:id":      "p1",
			"sys:flavour": "affine:paragraph",
			"prop:text":   "Broken leaf",
		},
	})

	if result.OK {
		t.Fatalf("OK = true, want false")
	}
	if result.Summary["missing_children"] != 1 {
		t.Fatalf("missing_children = %d, want 1", result.Summary["missing_children"])
	}
	if len(result.Issues) != 1 || result.Issues[0].ID != "p1" {
		t.Fatalf("issues = %#v, want one issue for p1", result.Issues)
	}
}

func TestCheckBlocksIntegrityAcceptsValidLeaf(t *testing.T) {
	result := CheckBlocksIntegrity("doc-1", map[string]map[string]any{
		"page": {
			"sys:id":       "page",
			"sys:flavour":  "affine:page",
			"sys:children": []any{"p1"},
		},
		"p1": {
			"sys:id":       "p1",
			"sys:flavour":  "affine:paragraph",
			"sys:children": []any{},
			"prop:text":    "Valid leaf",
		},
	})

	if !result.OK {
		t.Fatalf("OK = false, issues = %#v", result.Issues)
	}
	if result.IssueCount != 0 {
		t.Fatalf("IssueCount = %d, want 0", result.IssueCount)
	}
}

func TestCheckBlocksIntegrityDetectsMissingChildRef(t *testing.T) {
	result := CheckBlocksIntegrity("doc-1", map[string]map[string]any{
		"page": {
			"sys:id":       "page",
			"sys:flavour":  "affine:page",
			"sys:children": []any{"missing"},
		},
	})

	if result.OK {
		t.Fatalf("OK = true, want false")
	}
	if result.Summary["missing_child_ref"] != 1 {
		t.Fatalf("missing_child_ref = %d, want 1", result.Summary["missing_child_ref"])
	}
}

func TestCheckBlocksIntegrityDetectsConnectorBlocks(t *testing.T) {
	result := CheckBlocksIntegrity("doc-1", map[string]map[string]any{
		"page": {
			"sys:id":       "page",
			"sys:flavour":  "affine:page",
			"sys:children": []any{"bad-connector"},
		},
		"bad-connector": {
			"sys:id":       "bad-connector",
			"sys:flavour":  "affine:connector",
			"sys:children": []any{},
		},
	})

	if result.OK {
		t.Fatalf("OK = true, want false")
	}
	if result.Summary["unsupported_connector_block"] != 1 {
		t.Fatalf("unsupported_connector_block = %d, want 1", result.Summary["unsupported_connector_block"])
	}
}

func TestRepairIssueIDsSelectsConnectorBlocks(t *testing.T) {
	result := DocIntegrityResult{Issues: []DocIntegrityIssue{
		{Code: "missing_children", ID: "paragraph"},
		{Code: "unsupported_connector_block", ID: "connector"},
	}}

	got := repairIssueIDs(result, "connector-blocks")
	if len(got) != 1 || got[0] != "connector" {
		t.Fatalf("ids = %#v, want connector", got)
	}
}

func TestFilterRepairIDsKeepsOnlySelectedCandidates(t *testing.T) {
	got := filterRepairIDs([]string{"a", "b", "c"}, []string{" c ", "missing", "a"})
	want := []string{"a", "c"}
	if len(got) != len(want) {
		t.Fatalf("ids = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ids = %#v, want %#v", got, want)
		}
	}
}

func TestValidateRepairApplyRequiresExplicitConnectorBlockScope(t *testing.T) {
	if err := validateRepairApply("connector-blocks", nil, false, 2); err == nil {
		t.Fatal("error = nil, want explicit scope error")
	}
	if err := validateRepairApply("connector-blocks", []string{"a"}, false, 2); err != nil {
		t.Fatalf("targeted repair error: %v", err)
	}
	if err := validateRepairApply("connector-blocks", nil, true, 2); err != nil {
		t.Fatalf("broad repair error: %v", err)
	}
}
