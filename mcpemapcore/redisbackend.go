// redisbackend.go
package mcpemapcore

import (
	"fmt"
	"log"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

type RedisBackend struct {
}

func (r RedisBackend) UpdateMap(user *User,
	mapid int,
	uploadFilename string,
	pureHash string,
) {
	var err error
	_, err = conn.Do("HSET", "map:"+strconv.Itoa(mapid), "mapfilehash", pureHash)
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Do("HSET", "mapfilehash:"+pureHash, "id", strconv.Itoa(mapid))
	if err != nil {
		log.Fatal(err)
	}

	var nextGood int
	nextGood, err = redis.Int(conn.Do("INCR", "next_good"))
	if err != nil {
		log.Fatal(err)
	}
	_, err = conn.Do("ZADD", "goodmapset", nextGood, mapid)
	if err != nil {
		log.Fatal(err)
	}
	badid, err := redis.Int(conn.Do("ZSCORE", "badmapset", mapid))
	if err != nil {
		log.Fatal(err)
	}
	//nextBad, err = redis.Int(conn.Do("INCR", "next_bad"))
	_, err = conn.Do("ZREM", "badmapset", badid, mapid)
	if err != nil {
		log.Fatal(err)
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
