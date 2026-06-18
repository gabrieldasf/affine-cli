package canvaswrite

import (
	"os"

	"affine-pp-cli/internal/socketio"
)

func joinWorkspace(client *socketio.Client, workspaceID string) error {
	version := os.Getenv("AFFINE_WS_CLIENT_VERSION")
	if version == "" {
		version = "0.26.0"
	}
	return client.JoinWorkspace(workspaceID, version)
}
