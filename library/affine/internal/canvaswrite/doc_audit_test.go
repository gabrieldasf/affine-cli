package canvaswrite

import (
	"strings"
	"testing"

	"affine-pp-cli/internal/config"
)

func TestAuditDocSnapshotDoesNotRequireWorkspace(t *testing.T) {
	_, err := AuditDoc(&config.Config{}, DocAuditOptions{SnapshotFile: "missing-snapshot.bin"})
	if err == nil {
		t.Fatal("AuditDoc error = nil, want missing snapshot error")
	}
	if strings.Contains(err.Error(), "--workspace") {
		t.Fatalf("AuditDoc error = %q, should not require workspace for snapshot-file", err)
	}
}

func TestCheckDocIntegritySnapshotDoesNotRequireWorkspace(t *testing.T) {
	_, err := CheckDocIntegrity(&config.Config{}, DocIntegrityOptions{SnapshotFile: "missing-snapshot.bin"})
	if err == nil {
		t.Fatal("CheckDocIntegrity error = nil, want missing snapshot error")
	}
	if strings.Contains(err.Error(), "--workspace") {
		t.Fatalf("CheckDocIntegrity error = %q, should not require workspace for snapshot-file", err)
	}
}
