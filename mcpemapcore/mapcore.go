package mcpemapcore

import (
	"crypto/md5"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

const (
	address = "127.0.0.1:6379"
)

var (
	conn redis.Conn
)

type User struct {
	Id       string
	Username string `redis:"username"`
	Auth     string `redis:"auth"`
}

type Map struct {
	Id             string
	MapTitle       string `redis:"map_title"`
	Description    string `redis:"description"`
	MapDownloadUri string `redis:"mapdownloaduri"`
	MapFileHash    string `redis:"mapfilehash"`
	Author         string `redis:"author"`
	AuthorUri      string `redis:"author_uri"`
	NumViews       int    `redis:"numviews"`
	Tested         bool   `redis:"tested"`
	Featured       bool   `redis:"featured"`
	DownloadCount  int64  `redis:"downloadcount"`

	MapImageUriList []*MapImage
}

type MapImage struct {
	MapImageUri  string `redis:"mapimageuri"`
	MapImageHash string `redis:"mapimagehash"`
}

type Stats struct {
	Total_tested        int64 `redis:"total_tested"`
	Total_bad_tested    int64 `redis:"total_bad_tested"`
	Total_bad           int64 `redis:"total_bad"`
	Total_maps          int64 `redis:"total_maps"`
	Total_bad_featured  int64 `redis:"total_bad_featured"`
	Total_good_tested   int64 `redis:"total_good_tested"`
	Total_good_featured int64 `redis:"total_good_featured"`
	Total_good          int64 `redis:"total_good"`
}

func init() {
	var err error
	conn, err = redis.Dial("tcp", address)
	if nil != err {
		log.Fatalln("Error: Connection to redis:", err)
	}
	if !exists("maps") {
		err = os.Mkdir("maps", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
	if !exists("mapimages") {
		err = os.Mkdir("mapimages", 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func exists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		} else {
			log.Fatal(err)
		}
	}
	return true
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

func UpdateMapDownloadCount(fileHash string) error {
	var err error
	var mapId string
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

func WriteImageUriList(postId int, mapList []interface{}) error {
	var err error
	for _, i := range mapList {
		actual := i.(map[string]interface{})["MapImageUri"]
		success, hash := DownloadContent(actual.(string), "mapimages", "", "jpeg")
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
				"mapimagehash", hash)

			_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), mapImageId)
		}
		if err != nil {
			return err
		}
	}
	return err
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
	u.MapImageUri = fmt.Sprintf("%v/mapimages/%v.jpeg", siteRoot, u.MapImageHash)
	return u, nil
}

func GetAllMapsFromRedis(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "goodmapset")
}

func GetFeaturedMapsFromRedis(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "featuredmapset")
}

func GetMostDownloadedMapsFromRedis(start, count int64, siteRoot string) ([]*Map, int64, error) {
	return GetMapsFromRedis(start, count, siteRoot, "mostdownloaded")
}

func GetMapsFromRedis(start, count int64, siteRoot string, keyName string) ([]*Map, int64, error) {
	values, err := redis.Strings(conn.Do("ZRANGE", keyName, start, start+count-1))
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

func DownloadContent(uri string, dir string, acceptMime string, ext string) (bool, string) {
	resp, err := http.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	if headerType == acceptMime || acceptMime == "" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Printf("Bad MimeType:%v\n", headerType)
	}
	return false, ""
}

func DownloadContentRedirect(uri string, dir string, acceptMime string, ext string) (bool, string) {
	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			fmt.Printf("Redirect from %v to %v\n", r.URL.Opaque, r.URL.Path)
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(uri)

	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	if headerType == acceptMime || acceptMime == "" {
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Printf("Bad MimeType:%v\n", headerType)
	}
	return false, ""
}

func DownloadContentRedirectSearch(uri string, dir string, acceptMime string, ext string) (bool, string) {

	check := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			fmt.Printf("Redirect from %v to %v\n", r.URL.Opaque, r.URL.Path)
			r.URL.Opaque = r.URL.Path
			return nil
		},
	}

	resp, err := check.Get(uri)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	headerType := resp.Header.Get("Content-Type")
	fmt.Printf("ContentType={0}\n", headerType)
	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	if headerType == acceptMime || acceptMime == "" {
		fn := md5.Sum([]byte(uri))
		filename := fmt.Sprintf("%x.%v", fn, ext)
		hash := fmt.Sprintf("%x", fn)
		filepath := fmt.Sprintf("%v/%v", dir, filename)
		fmt.Println(filepath)
		err = ioutil.WriteFile(filepath, bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Println("HTML detected.. searching")
		//fmt.Printf("%v", string(bytes))
		searchFor :=
			`(http://[a-z_\/0-9\-\#=&\.|,|;|\?|\!]*/.zip)`
		stepRegex, _ := regexp.Compile(searchFor)
		captures := stepRegex.FindStringSubmatch(string(bytes))
		if len(captures) > 0 {
			fmt.Printf("Found:%v\n", len(captures))
			for i := range captures {
				fmt.Printf("Capture: %v %v\n", i, captures[i])

			}
		} else {
			fmt.Printf("Not found")
		}

	}
	return false, ""
}

func LoadUserInfo(userId string) (*User, error) {
	v, err := redis.Values(conn.Do("HGETALL", "user:"+userId))
	if err != nil {
		return nil, err
	}
	u := &User{Id: userId}
	err = redis.ScanStruct(v, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateFavoriteMap(u *User, mapId string, fav bool) error {

	var err error
	//add mapid to favorite set for user
	if fav {
		_, err = redis.Int(conn.Do("SADD", "favorite:"+u.Id, mapId))
		if err != nil {
			return err
		}

		//incrememnt favorite count on map:w
		_, err = redis.Int(conn.Do("HINCRBY", "map:"+mapId, "favoritecount", 1))
		if err != nil {
			log.Fatal(err)
		}
	} else {
		_, err = redis.Int(conn.Do("SPOP", "favorite:"+u.Id))
		if err != nil {
			return err
		}

		//decrement favorite count on map:w
		_, err = redis.Int(conn.Do("HINCRBY", "map:"+mapId, "favoritecount", -1))
		if err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func GetFavoriteMaps(u *User, siteRoot string) ([]*Map, error) {

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
