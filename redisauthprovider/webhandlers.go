package redisauth

import (
	"encoding/json"
	"errors"
	"fmt"
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

func RegisterUserRegistrationHandlerNoMux() {
	http.HandleFunc("/register", registerUser)
}

func RegisterUserRegistrationHandler(mux *http.ServeMux) {
	mux.HandleFunc("/register", registerUser)
}

func (a RedisUserProvider) login(username string, password string) (result bool, userid string) {
	err, id := login(username, password)
	if err != nil {
		fmt.Printf("Login:Error:%v", err.Error())
		return false, ""
	}

	fmt.Printf("Login:%v\n", username)
	return true, id
}

func registerUser(wr http.ResponseWriter, r *http.Request) {
	logger.SetPrefix("registerUser")
	wr.Header().Set("Access-Control-Allow-Origin", "*")
	type RegisterUser struct {
		Username string
		Password string
	}

	var b RegisterUser

	bodyBytes, err := ioutil.ReadAll(r.Body)
	logger.Printf("%s\n", bodyBytes)

	if hasFailed(wr, err) {
		return
	}

	err = json.Unmarshal(bodyBytes, &b)

	if hasFailed(wr, err) {
		logger.Printf("Error unmarshal")
		return
	}

	if b.Username == "" || b.Password == "" {
		hasFailed(wr, errors.New("Username or Password is empty or invalid."))
		return
	}

	_, err = register(b.Username, b.Password)
	logger.Println("user successfully created")

	if hasFailed(wr, err) {
		fmt.Printf("Error:" + err.Error())
		return
	}
}
