// redisbackend.go
package mcpemapcore

import (
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
