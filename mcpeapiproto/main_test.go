package main

import (
	"bytes"
	"clarkezone-vs-com/mcpemapcore"
	"clarkezone-vs-com/redisauthprovider"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/clarkezone/jwtauth"
)

var (
	auth *jwtauth.ApiSecurity
)

func TestGetMap(t *testing.T) {
	themap, err := mcpemapcore.GetMapFromRedis("1", "")
	if err != nil {
		log.Fatal(err)
	}
	_, err = json.Marshal(themap)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("%s", bytes)
}

func TestMain(m *testing.M) {
	var provider = redisauth.RedisUserProvider{}
	auth = jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlers()
	redisauth.RegisterUserRegistrationHandler()

	//TODO: make sure we don't write tests to db0
	//TODO: tests should work without preexisting redis state
	//	var result string
	//	var err error
	//	result, err = redis.String(conn.Do("SELECT", 1))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Printf("Swiched to db 1 for testing %v\n", result)
	os.Exit(m.Run())
}

func resetTestDb() {
	//	result, err := redis.String(conn.Do("FLUSHDB"))
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Printf("Reseting database\n", result)
}

func TestFavoriteMap(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(auth.RequireTokenAuthentication(UpdateFavoriteMap)))
	defer ts.Close()

	type Body struct {
		MapId string
		Add   bool
	}

	var b Body

	b.MapId = "1"
	b.Add = true
	testGet(ts, b, "1", t)
}

func TestGetFavories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(auth.RequireTokenAuthentication(GetMaps)))
	defer ts.Close()
	ts.URL += "/getuserfavorites"
	testGet(ts, nil, "1", t)
}
