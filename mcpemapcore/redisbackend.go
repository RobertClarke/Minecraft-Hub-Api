// redisbackend.go
package mcpemapcore

import (
	"errors"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

const (
	address = "127.0.0.1:6379"
)

var (
	conn redis.Conn
)

func init() {
}

type RedisBackend struct {
	currentDatabase int
	logger          *log.Logger
}

func CreateRedisBackendWithDatabase(db int) *RedisBackend {
	redisInstance := RedisBackend{}
	redisInstance.currentDatabase = db

	//redisInstance.logger = (log.New(ioutil.Discard, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile))
	redisInstance.logger = (log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile))

	var err error
	conn, err = redis.Dial("tcp", address, redis.DialDatabase(redisInstance.currentDatabase))
	if nil != err {
		log.Fatalln("Error: Connection to redis:", err)
	}
	fmt.Printf("redis is alive\n")

	return &redisInstance
}

func CreateRedisBackend() *RedisBackend {
	return CreateRedisBackendWithDatabase(0)
}

func (r RedisBackend) CreateMap(user *User,
	newMap *NewMap) (string, error) {

	r.logger.Println("RedisBackend:CreateMap")

	dir, _ := os.Getwd()
	mapDir := path.Join(dir, "maps")
	filePath := path.Join(mapDir, newMap.MapFilename+".zip")
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		r.logger.Printf("FAILED Create map: file %v\n", newMap.MapFilename)
		return "", errors.New("map doesn't exist")
	} else {
		r.logger.Printf("Create map: file %v exists\n", newMap.MapFilename)
	}

	imageDir := path.Join(dir, "mapimages")

	theNewMap := Map{
		MapTitle:    newMap.Title,
		Description: newMap.Description,
		//MapFileHash:     sanitize(newMap.MapFilename),
		MapFileHash:     newMap.MapChecksum,
		MapImageUriList: make([]*MapImage, len(newMap.MapImageFileNames)),
	}

	for i := range newMap.MapImageFileNames {
		iFn := path.Join(imageDir, newMap.MapImageFileNames[i])
		r.logger.Printf("verifying %v", iFn)
		_, err := os.Stat(iFn)
		if os.IsNotExist(err) {
			r.logger.Printf("FAILED Create map: imagefile %v\n", iFn)
			return "", errors.New("map image doesn't exist " + iFn)
		} else {
			r.logger.Printf("Create map: imagefile %v exists\n", iFn)
		}

		filename := fmt.Sprintf("%v%v", newMap.MapImageChecksums[i], path.Ext(newMap.MapImageFileNames[i]))
		//Add the URI here
		var mi MapImage = MapImage{}
		mi.MapImageHash = newMap.MapChecksum
		mi.MapImageUri = mi.MapImageHash
		mi.MapImageFilename = filename

		theNewMap.MapImageUriList[i] = &mi
	}

	//writeMapFromMap(&theNewMap)
	postId, err := redis.Int(conn.Do("INCR", "next_map_id"))
	if err != nil {
		return "", err
	}
	err = writeMapFromMap(postId, &theNewMap, true)
	return "", err
}

func sanitize(name string) string {
	res := strings.Split(".", name)
	return res[0]
}

func (r RedisBackend) GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "goodmapset", false)
}

func (r RedisBackend) GetFeaturedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "featuredmapset", false)
}

func (r RedisBackend) GetMostDownloadedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "mostdownloaded", true)
}

func (r RedisBackend) GetMostFavoritedMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "mostfavorited", true)
}

func (r RedisBackend) GetFavoriteMaps(u *User, start, count int64, siteRoot string) ([]*Map, int64, error) {
	panic("not implemented")
}

func (r RedisBackend) LoadUserInfo(userid string) (*User, error) {

	return RedisLoadUserInfo(userid)
}

func (r RedisBackend) UpdateFavoriteMap(u *User, mapId string, fav bool) error {
	return RedisUpdateFavoriteMap(u, mapId, fav)
}

func (r RedisBackend) UpdateMapDownloadCount(hash string) error {
	return RedisUpdateMapDownloadCount(hash)
}

func (r RedisBackend) UpdateMap(user *User,
	mapid int,
	uploadFilename string,
	pureHash string,
) {
	var err error
	_, err = conn.Do("HSET", "map:"+strconv.Itoa(mapid), "mapfilehash", pureHash)
	if err != nil {
		r.logger.Fatal(err)
	}
	_, err = conn.Do("HSET", "mapfilehash:"+pureHash, "id", strconv.Itoa(mapid))
	if err != nil {
		r.logger.Fatal(err)
	}

	var nextGood int
	nextGood, err = redis.Int(conn.Do("INCR", "next_good"))
	if err != nil {
		r.logger.Fatal(err)
	}
	_, err = conn.Do("ZADD", "goodmapset", nextGood, mapid)
	if err != nil {
		r.logger.Fatal(err)
	}
	badid, err := redis.Int(conn.Do("ZSCORE", "badmapset", mapid))
	if err != nil {
		r.logger.Fatal(err)
	}
	//nextBad, err = redis.Int(conn.Do("INCR", "next_bad"))
	_, err = conn.Do("ZREM", "badmapset", badid, mapid)
	if err != nil {
		r.logger.Fatal(err)
	}
}

func RedisUpdateMapDownloadCount(fileHash string) error {
	var err error
	var mapId string
	fmt.Printf("Download Hash:%v\n", fileHash)
	mapId, err = redis.String(conn.Do("HGET", "mapfilehash:"+fileHash, "id"))
	if err != nil {
		log.Fatal(err)
	}

	if mapId != "" {
		fmt.Printf("found map id %v for file %v\n", mapId, fileHash)
		var count int
		count, err = redis.Int(conn.Do("HINCRBY", "map:"+mapId, "downloadcount", 1))
		if err != nil {
			log.Fatal(err)
		}
		_, err = redis.Int(conn.Do("ZADD", "mostdownloaded", count, mapId))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("NOT found map id for file %d", fileHash)
	}
	return err
}
func RedisUpdateFavoriteMap(u *User, mapId string, fav bool) error {

	var err error
	var existingCount int
	//add mapid to favorite set for user
	if fav {
		_, err = redis.Int(conn.Do("SADD", "favorite:"+u.Id, mapId))
		if err != nil {
			return err
		}

		//incrememnt favorite count on map:w
		existingCount, err = redis.Int(conn.Do("HINCRBY", "map:"+mapId, "favoritecount", 1))
		if err != nil {
			log.Fatal(err)
		}
		_, err = redis.Int(conn.Do("ZADD", "mostfavorited", existingCount, mapId))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err = redis.Int(conn.Do("SPOP", "favorite:"+u.Id))
		if err != nil {
			return err
		}

		//decrement favorite count on map:w
		existingCount, err = redis.Int(conn.Do("HINCRBY", "map:"+mapId, "favoritecount", -1))
		if err != nil {
			log.Fatal(err)
		}
		_, err = redis.Int(conn.Do("ZADD", "mostfavorited", existingCount, mapId))
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}
func RedisGetFavoriteMaps(u *User, siteRoot string) ([]*Map, error) {

	values, err := redis.Strings(conn.Do("SMEMBERS", "favorite:"+u.Id))
	if err != nil {
		return nil, err
	}
	maps := []*Map{}
	for _, mid := range values {
		m, err := GetMapFromRedis(mid, siteRoot)
		if err == nil {
			maps = append(maps, m)
		}
	}
	return maps, nil
}

func GetStats() Stats {
	v, err := redis.Values(conn.Do("HGETALL", "stats"))
	if err != nil {
		log.Fatal(err)
	}
	u := &Stats{}
	err = redis.ScanStruct(v, u)
	if err != nil {
		log.Fatal(err)
	}
	return *u
}

func WriteNextMap(object map[string]interface{}, good bool, mapfilehash string) error {
	postId, err := redis.Int(conn.Do("INCR", "next_map_id"))
	if err != nil {
		return err
	}
	err = writeMap(postId, object, good, mapfilehash)

	return err
}

func KillwriteMapFromMap(m *Map) error {
	//verify mapfilehash is not null
	//verify downloaduri is not null

	postId, err := redis.Int(conn.Do("INCR", "next_map_id"))
	if err != nil {
		return err
	}
	log.Printf("Writing map to redis for id:%v filehash:%v", postId, m.MapFileHash)
	_, err = conn.Do("HMSET",
		fmt.Sprintf("map:%d", postId),
		"map_title", m.MapTitle,
		"description", m.Description,
		"author", m.Author,
		"author_uri", m.AuthorUri,
		"mapdownloaduri", m.MapDownloadUri,
		"mapfilehash", m.MapFileHash,
		"numviews", 0,
		"tested", 1,
		"featured", 1,
		"time", time.Now().Unix())

	if err != nil {
		return err
	}
	_, err = conn.Do("HMSET",
		"mapfilehash:"+m.MapFileHash,
		"id", fmt.Sprintf("%d", postId))
	if err != nil {
		return err
	}
	log.Println("Map written")
	return nil
}

func writeMapFromMap(postId int, m *Map, good bool) error {
	var err error
	var nextGood, nextBad, nextTested, nextFeatured int
	_, err = conn.Do("HMSET",
		fmt.Sprintf("map:%d", postId),
		"map_title", m.MapTitle,
		"description", m.Description,
		"author", m.Author,
		"author_uri", m.AuthorUri,
		"mapdownloaduri", m.MapDownloadUri,
		"mapfilehash", m.MapFileHash,
		"numviews", 0,
		"tested", 1,
		"featured", 1,
		"time", time.Now().Unix())

	if err != nil {
		return err
	}

	_, err = conn.Do("HMSET",
		"mapfilehash:"+m.MapFileHash,
		"id", fmt.Sprintf("%d", postId))

	if err != nil {
		return err
	}

	if m.Tested == true {
		if good {
			nextTested, err = redis.Int(conn.Do("INCR", "next_tested"))
			conn.Do("ZADD", "testedmapset", nextTested, postId)
		}
		//conn.Do("LPUSH", "testedmaplist", postId)
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_tested", 1))
		if good {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good_tested", 1))
		} else {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad_tested", 1))
		}
	}
	if m.Featured == true {
		if good {
			nextFeatured, err = redis.Int(conn.Do("INCR", "next_featured"))
			conn.Do("ZADD", "featuredmapset", nextFeatured, postId)
		}
		//conn.Do("LPUSH", "featuredmaplist", postId)
		_, err = redis.Int(conn.Do("INCR", "total_featured"))
		if good {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good_featured", 1))
		} else {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad_featured", 1))
		}
	}
	if good {
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good", 1))
	} else {
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad", 1))
	}
	_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_maps", 1))
	if err != nil {
		return err
	}

	err = WriteImageUriListNoDownload(postId, m.MapImageUriList)
	if err != nil {
		return err
	}

	if good {
		nextGood, err = redis.Int(conn.Do("INCR", "next_good"))
		conn.Do("ZADD", "goodmapset", nextGood, postId)
	} else {
		nextBad, err = redis.Int(conn.Do("INCR", "next_bad"))
		conn.Do("ZADD", "badmapset", nextBad, postId)
	}
	if err != nil {
		return err
	}

	//uriList := object["MapUriList"].([]map[string]interface{})

	return err
}
func writeMap(postId int, object map[string]interface{}, good bool, mapfilehash string) error {
	var err error
	var nextGood, nextBad, nextTested, nextFeatured int
	_, err = conn.Do("HMSET",
		fmt.Sprintf("map:%d", postId),
		"map_title", object["MapTitle"],
		"description", object["Description"],
		"author", object["Author"],
		"author_uri", object["AuthorUri"],
		"mapdownloaduri", object["MapDownloadUri"],
		"mapfilehash", mapfilehash,
		"numviews", object["NumViews"],
		"tested", object["Tested"],
		"featured", object["Featured"],
		"time", time.Now().Unix())

	_, err = conn.Do("HMSET",
		"mapfilehash:"+mapfilehash,
		"id", fmt.Sprintf("%d", postId))

	if object["Tested"] == "1" {
		if good {
			nextTested, err = redis.Int(conn.Do("INCR", "next_tested"))
			conn.Do("ZADD", "testedmapset", nextTested, postId)
		}
		//conn.Do("LPUSH", "testedmaplist", postId)
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_tested", 1))
		if good {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good_tested", 1))
		} else {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad_tested", 1))
		}
	}
	if object["Featured"] == "1" {
		if good {
			nextFeatured, err = redis.Int(conn.Do("INCR", "next_featured"))
			conn.Do("ZADD", "featuredmapset", nextFeatured, postId)
		}
		//conn.Do("LPUSH", "featuredmaplist", postId)
		_, err = redis.Int(conn.Do("INCR", "total_featured"))
		if good {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good_featured", 1))
		} else {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad_featured", 1))
		}
	}
	if good {
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good", 1))
	} else {
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad", 1))
	}
	_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_maps", 1))
	if err != nil {
		return err
	}

	uriList := object["MapImageUriList"].([]interface{})

	WriteImageUriList(postId, uriList)

	//	for _, i := range uriList {
	//		actual := i.(map[string]interface{})["MapImageUri"]
	//		_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), actual)
	//		if err != nil {
	//			return err
	//		}
	//	}

	if good {
		nextGood, err = redis.Int(conn.Do("INCR", "next_good"))
		conn.Do("ZADD", "goodmapset", nextGood, postId)
		//conn.Do("LPUSH", "goodmaplist", postId)
	} else {
		nextBad, err = redis.Int(conn.Do("INCR", "next_bad"))
		conn.Do("ZADD", "badmapset", nextBad, postId)
		//conn.Do("LPUSH", "badmaplist", postId)
	}
	if err != nil {
		return err
	}

	//uriList := object["MapUriList"].([]map[string]interface{})

	return err
}

func WriteImageUriListNoDownload(postId int, maps []*MapImage) error {
	log.Printf("writing map image list %v\n", len(maps))
	var err error
	for i := range maps {
		actual := *(maps[i])
		mapImageId, err := redis.Int(conn.Do("INCR", "next_mapImage_id"))
		if err != nil {
			return err
		}
		_, err = conn.Do("HMSET",
			fmt.Sprintf("mapimage:%d", mapImageId),
			"mapimageuri", actual.MapImageUri,
			"mapimagehash", actual.MapImageHash,
			"mapimagefilename", actual.MapImageFilename)

		if err != nil {
			return err
		}

		_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), mapImageId)
		if err != nil {
			return err
		}
	}
	return err
}

func WriteImageUriList(postId int, mapList []interface{}) error {
	var err error
	for _, i := range mapList {
		actual := i.(map[string]interface{})["MapImageUri"]
		success, hash := DownloadContent(actual.(string), "mapimages", "", path.Ext(actual.(string)))
		filename := fmt.Sprintf("%v%v", hash, path.Ext(actual.(string)))
		//		success := true
		//		hash := ""
		if success {
			mapImageId, err := redis.Int(conn.Do("INCR", "next_mapImage_id"))
			if err != nil {
				return err
			}
			_, err = conn.Do("HMSET",
				fmt.Sprintf("mapimage:%d", mapImageId),
				"mapimageuri", actual,
				"mapimagehash", hash,
				"mapimagefilename", filename)

			_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), mapImageId)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func GetMapsFromRedis(start, count int64, siteRoot string, keyName string, reverse bool) ([]*Map, int64, error) {
	var values []string
	var err error
	if reverse {
		values, err = redis.Strings(conn.Do("ZREVRANGE", keyName, start, start+count-1))
	} else {
		values, err = redis.Strings(conn.Do("ZRANGE", keyName, start, start+count-1))
	}
	if err != nil {
		return nil, 0, err
	}
	maps := []*Map{}
	for _, mid := range values {
		m, err := GetMapFromRedis(mid, siteRoot)
		if err == nil {
			maps = append(maps, m)
		}
	}
	r, err := redis.Int64(conn.Do("ZCARD", "maplist"))
	if err != nil {
		return maps, 0, nil
	} else {
		return maps, r - start - int64(len(values)), nil
	}
}

func GetMapFromRedis(mapId string, siteRoot string) (*Map, error) {
	v, err := redis.Values(conn.Do("HGETALL", "map:"+mapId))
	if err != nil {
		log.Fatal(err)
	}
	u := &Map{Id: mapId}
	err = redis.ScanStruct(v, u)
	if err != nil {
		log.Fatal(err)
	}

	u.MapOriginalUri = u.MapDownloadUri
	if len(siteRoot) >= 7 && strings.ToLower(siteRoot) != "http://" {
		siteRoot = "http://" + siteRoot
	}
	//TODO: enable this rewriting of download uri to be disabled and configured.
	//u.MapDownloadUri = fmt.Sprintf("http://%v/maps/%v.zip", siteRoot, u.MapFileHash)
	u.MapDownloadUri = fmt.Sprintf("%v/maps/%v.zip", siteRoot, u.MapFileHash)

	//Enumerate and gather mapimages

	u.MapImageUriList = GetMapImages(mapId, siteRoot)

	return u, nil
}

func GetMapImages(mapId string, siteRoot string) []*MapImage {
	mapImages := []*MapImage{}
	imageListKey := "mapimages:" + mapId
	len, err := redis.Int64(conn.Do("LLEN", imageListKey))
	if err != nil {
		log.Fatal(err)
	}

	values, err := redis.Strings(conn.Do("LRANGE", imageListKey, 0, len-1))

	for i := range values {
		m, err := GetMapImageFromRedis(values[i], siteRoot)
		if err == nil {
			mapImages = append(mapImages, m)
		}
	}
	return mapImages
}

func GetMapImageFromRedis(mapImageId string, siteRoot string) (*MapImage, error) {
	v, err := redis.Values(conn.Do("HGETALL", "mapimage:"+mapImageId))
	if err != nil {
		return nil, err
	}
	u := &MapImage{}
	err = redis.ScanStruct(v, u)
	if err != nil {
		return nil, err
	}
	//u.MapImageUri = fmt.Sprintf("http://%v/mapimages/%v.jpeg", siteRoot, u.MapImageHash)
	if len(siteRoot) >= 7 && strings.ToLower(siteRoot[:7]) != "http://" {
		siteRoot = "http://" + siteRoot
	}
	u.MapImageUri = fmt.Sprintf("%v/mapimages/%v", siteRoot, u.MapImageFilename)
	return u, nil
}
