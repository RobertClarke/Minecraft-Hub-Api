package mcpemapcore

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"testing"

	jwtauth "github.com/clarkezone/jwtauth-go"
	redisauth "github.com/robertclarke/Minecraft-Hub-Api/redisauthprovider"
)

var (
	auth *jwtauth.ApiSecurity
)

func TestMain(m *testing.M) {
	var provider = redisauth.RedisUserProvider{}
	auth = jwtauth.CreateApiSecurity(provider)
	auth.RegisterLoginHandlers()
	redisauth.RegisterUserRegistrationHandlerNoMux()

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

func TestCreateMap(t *testing.T) {
	logger := log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile)
	ts := httptest.NewServer(http.HandlerFunc(auth.RequireTokenAuthentication(HandleCreateMap)))
	defer ts.Close()

	request := createMapRequest{}
	request.Map = NewMap{}
	request.Map.Title = "Hello Map"
	request.Map.Description = "Hello Description"

	ts.URL += "/getuserfavorites"
	res := testGet(ts, nil, "1", t)
	if res.Code != http.StatusBadRequest || res.Error != "unexpected end of JSON input" {
		log.Fatal("expected bad JSON:", res.Error)
	}

	request.Map.MapImageFileNames = []string{"image1.png", "image2.png"}

	res = testGet(ts, request, "1", t)
	if res.Code != http.StatusBadRequest || res.Error != "map must have a filename" {
		log.Fatal("expected malformed payload:", res.Error)
	}

	request.Map.MapFilename = "m4sBABVgAAA=.zip"

	testDir, err := setupTestFiles(&request.Map, logger)
	if err != nil {
		log.Fatal(err.Error())
	}

	var mapBytes []byte
	mapBytes, err = ioutil.ReadFile(path.Join(testDir, request.Map.MapFilename))
	if err != nil {
		logger.Fatal("Test couldn't check checksum")
	}
	chkSum := md5.Sum(mapBytes)
	sh := fmt.Sprintf("%x", chkSum)
	request.Map.MapChecksum = sh
	logger.Printf("getting hash for map %v\n", sh)

	res = testGet(ts, request, "1", t)
	if res.Code != http.StatusOK {
		log.Fatal("expected: malformed payload got:", res.Error)
	}
	//TODO Checksum maps and images
}

func testGet(ser *httptest.Server, param interface{}, userid string, t *testing.T) createMapResponse {
	var currentAuth jwtauth.JwtAuthProvider

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
	res, _ := ioutil.ReadAll(result.Body)
	responsePayload := createMapResponse{}
	json.Unmarshal(res, &responsePayload)
	return responsePayload
}
