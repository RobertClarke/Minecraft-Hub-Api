package main

import (
	"bytes"
	"clarkezone-vs-com/mcpemapcore"
	"encoding/json"
	"github.com/clarkezone/jwtauth"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
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
	ts := httptest.NewServer(http.HandlerFunc(jwtauth.RequireTokenAuthentication(UpdateFavoriteMap)))
	defer ts.Close()

	type Body struct {
		MapId string
		Add   bool
	}

	var b Body

	b.MapId = "1"
	b.Add = true
	testGet(ts, b, t)
}

func TestGetFavories(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(jwtauth.RequireTokenAuthentication(GetMaps)))
	defer ts.Close()
	ts.URL += "/getuserfavorites"
	testGet(ts, nil, t)
}

func testGet(ser *httptest.Server, param interface{}, t *testing.T) {
	var currentAuth jwtauth.JwtAuthProvider

	userid := "1"
	token, _ := currentAuth.GenerateToken(userid)
	client := &http.Client{}

	var req *http.Request

	if param != nil {
		jsonBytes, err := json.Marshal(param)

		if err != nil {
			log.Fatal(err)
		}
		req, _ = http.NewRequest("GET", ser.URL, bytes.NewBuffer(jsonBytes))
	} else {
		req, _ = http.NewRequest("GET", ser.URL, nil)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	result, _ := client.Do(req)
	if result.StatusCode != http.StatusOK {
		t.Fail()
	}
}
