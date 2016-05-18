package mcpemapcore

import "testing"

func TestUploadMapDownloadCount(t *testing.T) {
	sql := MySqlBackend{}
	sql.UpdateMapDownloadCount("b22bf60c3bf35994c606059a59ed1f99")
}

func TestUpdateFavoriteMap(t *testing.T) {
	u := &User{}
	var mapId string
	mapId = "3"
	u.Id = "123"

	sql := MySqlBackend{}
	sql.UpdateFavoriteMap(u, mapId, true)
}
