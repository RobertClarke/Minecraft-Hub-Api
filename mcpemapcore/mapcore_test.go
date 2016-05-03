package mcpemapcore

import (
	"fmt"
	"testing"
)

func TestGetZipThruProxy(t *testing.T) {
	name := "http://www.mediafire.com/?p3ijl1h0s4iy10s"
	success, hash := DownloadContentRedirect(name, "maps", "application/zip", "zip")
	fmt.Printf("%v %v\n", hash, success)

}
func TestMySqlGetAllMaps(t *testing.T) {
	maps, err := MySqlGetAllMaps(0, 0, "")
	if err != nil {
		t.Fail()
	}
	fmt.Printf("got %v\n", len(maps))
}
