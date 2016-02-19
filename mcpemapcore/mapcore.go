package mcpemapcore

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"log"
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
	Author         string `redis:"author"`
	AuthorUri      string `redis:"author_uri"`
	NumViews       int    `redis:"numviews"`
	Tested         bool   `redis:"tested"`
	Featured       bool   `redis:"featured"`
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

	for _, i := range uriList {
		actual := i.(map[string]interface{})["MapImageUri"]
		_, err = conn.Do("LPUSH", "mapimages:"+strconv.Itoa(postId), actual)
		if err != nil {
			return err
		}
	}

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
