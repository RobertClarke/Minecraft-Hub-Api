package mcpemapcore

type Backend interface {
	CreateMap(user *User,
		newMap *NewMap) (string, error)

	UpdateMapDownloadCount(fileHash string) error

	UpdateFavoriteMap(u *User, mapId string, fav bool) error

	LoadUserInfo(userId string) (*User, error)

	UpdateMap(user *User, mapid int, uploadFilename string, hash string)

	GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error)

	GetFeaturedMaps(start, count int64, siteRoot string) ([]*Map, int64, error)

	GetMostDownloadedMaps(start, count int64, siteRoot string) ([]*Map, int64, error)

	GetMostFavoritedMaps(start, count int64, siteRoot string) ([]*Map, int64, error)

	GetFavoriteMaps(u *User, start, count int64, siteRoot string) ([]*Map, int64, error)
}
