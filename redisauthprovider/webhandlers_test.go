package redisauth

import (
	"bytes"
	"encoding/json"
	"github.com/clarkezone/jwtauth"
	"github.com/garyburd/redigo/redis"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	var result string
	var err error
	result, err = redis.String(conn.Do("SELECT", 1))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Swiched to db 1 for testing %v\n", result)
	os.Exit(m.Run())
}
func resetTestDb() {
	result, err := redis.String(conn.Do("FLUSHDB"))
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Reseting database\n", result)
}
func TestRegisterUser(t *testing.T) {
	log.Println("Testing register user")
	resetTestDb()
	ts := httptest.NewServer(http.HandlerFunc(registerUser))

	client := &http.Client{}

	type RegisterUser struct {
		Username string
		Password string
	}

	var b RegisterUser

	b.Username = "foo"
	b.Password = "bar"

	jsonBytes, err := json.Marshal(b)

	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest("GET", ts.URL, bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")

	result, _ := client.Do(req)
	if result.StatusCode != http.StatusOK {
		t.Fail()
	}
}

func TestLogin(t *testing.T) {
	log.Println("Testing user login")
	TestRegisterUser(t)
	var provider = redisUserProvider{}

	api := jwtauth.CreateApiSecurity(provider)

	ts := httptest.NewServer(http.HandlerFunc(api.Login))

	res, err := http.PostForm(ts.URL, url.Values{"username": {"foo"}, "password": {"bar"}})
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != http.StatusOK {
		t.Fail()
	}

	result := GetBody(res)

	user := jwtauth.UserFromToken(result)
	if user != "1" {
		t.Fail()
	}
}

func GetBody(res *http.Response) (result string) {
	response, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	return string(response)
}
