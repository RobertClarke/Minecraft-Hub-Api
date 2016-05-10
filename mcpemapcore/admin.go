package mcpemapcore

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type AdminMap struct {
	Id             string
	MapTitle       string `redis:"map_title"`
	Description    string `redis:"description"`
	MapDownloadUri string `redis:"mapdownloaduri"`
	MapOriginalUri string
	MapFileHash    string `redis:"mapfilehash"`
	Author         string `redis:"author"`
	AuthorUri      string `redis:"author_uri"`
	EditedById     int
	EditedByString string
	Edited         time.Time
	NumViews       int   `redis:"numviews"`
	Tested         bool  `redis:"tested"`
	Featured       bool  `redis:"featured"`
	DownloadCount  int64 `redis:"downloadcount"`
	FavoriteCount  int64 `redis:"favoritecount"`

	MapImageUriList []*MapImage
}

func MakeUserAdmin(username string) error {
	var err error
	user, err := LoadUserByUsername(username)

	if user != nil {
		user.AddToRole(*rolesByName["Administrator"])
		fmt.Printf("User %v is admin\n", user.Username)
	} else {
		fmt.Printf("User %v not found", username)
	}
	return err
}

func AdminUpdateMap(user *User, mapid int, uploadFilename string) error {

	var pureHash = strings.NewReplacer(".zip", "").Replace(uploadFilename)

	var err error
	if user != nil {
		sourceDir := "uploads/" + user.Username + "/" + uploadFilename
		destDir := "Maps/" + uploadFilename
		os.Rename(sourceDir, destDir)
	}
	if err != nil {
		return err
	}
	currentBackend.UpdateMap(user, mapid, uploadFilename, pureHash)
	return nil
}

func AdminGetBadMaps(start, count int64, siteRoot string) ([]*AdminMap, int64, error) {
	//return GetMapsFromRedis(start, count, siteRoot, "badmapset", true)
	maps, err := MySqlAdminGetBadMaps(int(start), int(count), siteRoot)
	return maps, -1, err
}

func AdminGetEditedMaps(start, count int64, siteRoot string) ([]*AdminMap, int64, error) {
	maps, err := MySqlAdminGetEditedMaps(int(start), int(count), siteRoot)
	return maps, -1, err
}
