package mcpemapcore

import "testing"

func TestUploadMapDownloadCount(t *testing.T) {
	sql := MySqlBackend{}
	sql.UpdateMapDownloadCount("b22bf60c3bf35994c606059a59ed1f99")
}
