package mcpemapcore

import (
	"crypto/md5"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

const (
	address = "127.0.0.1:6379"
)

var (
	conn redis.Conn
)

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
}

type MapImage struct {
	MapImageUri  string
	MapImageHash string
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

func writeMap(postId int, object map[string]interface{}, good bool, mapfilehash string) error {
	var err error
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
	if object["Tested"] == "1" {
		conn.Do("LPUSH", "testedmaplist", postId)
		_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_tested", 1))
		if good {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_good_tested", 1))
		} else {
			_, err = redis.Int(conn.Do("HINCRBY", "stats", "total_bad_tested", 1))
		}
	}
	if object["Featured"] == "1" {
		conn.Do("LPUSH", "featuredmaplist", postId)
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
		conn.Do("LPUSH", "goodmaplist", postId)
	} else {

		conn.Do("LPUSH", "badmaplist", postId)
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
		success, hash := DownloadContent(actual.(string), "mapimages", "")

		if success {
			mapImageId, err := redis.Int(conn.Do("INCR", "next_mapImage_id"))
			if err != nil {
				return err
			}
			_, err = conn.Do("HMSET",
				fmt.Sprintf("mapimage:%d", mapImageId),
				"mapimageuri", actual,
				"mapimageuash", hash)

			_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), mapImageId)
		}
		if err != nil {
			return err
		}
	}
	return err
}

func GetMapFromRedis(mapId string) (*Map, error) {
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
	u.MapDownloadUri = fmt.Sprintf("http://localhost:8080/maps/%v.zip", u.MapFileHash)

	return u, nil
}

func GetMapsFromRedis(start, count int64) ([]*Map, int64, error) {
	values, err := redis.Strings(conn.Do("LRANGE", "goodmaplist", start, start+count-1))
	if err != nil {
		return nil, 0, err
	}
	maps := []*Map{}
	for _, mid := range values {
		m, err := GetMapFromRedis(mid)
		if err == nil {
			maps = append(maps, m)
		}
	}
	r, err := redis.Int64(conn.Do("LLEN", "maplist"))
	if err != nil {
		return maps, 0, nil
	} else {
		return maps, r - start - int64(len(values)), nil
	}
}

func DownloadContent(uri string, dir string, acceptMime string) (bool, string) {
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
		filename := fmt.Sprintf("%x.zip", fn)
		hash := fmt.Sprintf("%x", fn)
		err = ioutil.WriteFile(fmt.Sprintf("%v/%v", dir, filename), bytes, os.FileMode(0777))

		if err != nil {
			log.Fatal(err)
		}
		return true, hash
	} else {
		fmt.Printf("Bad MimeType:%v\n", headerType)
	}
	return false, ""
}
