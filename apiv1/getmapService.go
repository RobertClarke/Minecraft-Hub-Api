package main

import "time"

// Map structure
type Map struct {
	ID             string
	MapTitle       string `redis:"map_title" db:"title"`
	Description    string `redis:"description"`
	MapDownloadURI string `redis:"mapdownloaduri"`
	MapOriginalURI string
	MapFileHash    string `redis:"mapfilehash"`
	Author         string `redis:"author"`
	AuthorURI      string `redis:"author_uri"`
	NumViews       int    `redis:"numviews"`
	Tested         bool   `redis:"tested"`
	Featured       bool   `redis:"featured"`
	DownloadCount  int64  `redis:"downloadcount" db:"downloads"`
	FavoriteCount  int64  `redis:"favoritecount"`

	MapImageURIList []*MapImage
}

// MapImage type
type MapImage struct {
	MapImageURI      string `redis:"mapimageuri"`
	MapImageHash     string `redis:"mapimagehash"`
	MapImageFilename string `redis:"mapimagefilename"`
}

// GetMapService interface
type GetMapService interface {
	GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error)
	EnsureDirectDL(postid int) error
}

type getMapService struct {
	myBackend Backend
}

// CreateGetMapService returns a new mapservice
func CreateGetMapService() GetMapService {
	return MySqlBackend{}
}

func (s getMapService) GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	metricstart := time.Now()
	defer func() {
		l := time.Since(metricstart)
		ms := float64(l.Nanoseconds() * 1000)
		apibackendlatencyms.Observe(ms)
	}()
	maps, next, err := s.myBackend.GetAllMaps(start, count, siteRoot)
	//return GetMapsFromRedis(start, count, siteRoot, "goodmapset", false)
	return maps, next, err
}
