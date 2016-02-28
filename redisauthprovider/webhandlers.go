package redisauth

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/clarkezone/jwtauth"
	"io/ioutil"
	"net/http"
)

type errorObj struct {
	Message string
}

func hasFailed(wr http.ResponseWriter, err error) bool {
	if err != nil {
		wr.WriteHeader(http.StatusInternalServerError)

		o := errorObj{err.Error()}

		buffer, _ := json.Marshal(o)

		wr.Write(buffer) //ignore errors

		return true
	}
	return false
}

func RegisterAuthHandlers() {
	var provider = redisUserProvider{}
	api := jwtauth.CreateApiSecurity(provider)
	api.RegisterAuthHandlers()
	http.HandleFunc("/register", registerUser)
}

func (a redisUserProvider) login(username string, password string) (result bool, userid string) {
	err, id := login(username, password)
	if err != nil {
		fmt.Printf("Login:Error:%v", err.Error())
		return false, ""
	}

	fmt.Printf("Login:%v\n", username)
	return true, id
}

func registerUser(wr http.ResponseWriter, r *http.Request) {
	type RegisterUser struct {
		Username string
		Password string
	}

	var b RegisterUser

	bodyBytes, err := ioutil.ReadAll(r.Body)

	if hasFailed(wr, err) {
		return
	}

	err = json.Unmarshal(bodyBytes, &b)

	if hasFailed(wr, err) {
		return
	}

	if b.Username == "" || b.Password == "" {
		hasFailed(wr, errors.New("Username or Password is empty or invalid."))
		return
	}

	_, err = register(b.Username, b.Password)

	if hasFailed(wr, err) {
		fmt.Printf("Error:" + err.Error())
		return
	}
}
