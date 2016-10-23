package redisauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/garyburd/redigo/redis"
)

func main() {

}

func init() {
}

var (
	key    = []byte("h27d72ts7hiasehi711334")
	logger = log.New(os.Stdout, "TRACE:", log.Ldate|log.Ltime|log.Lshortfile)
)

func register(username, password string) (auth string, err error) {
	userId, err := redis.Int(conn.Do("INCR", "next_user_id"))
	if err != nil {
		return "", err
	}
	hashedPw := hashPw(password)
	//auth = string(securecookie.GenerateRandomKey(32)) // We reuse the securecookie random string generator
	auth, err = redis.String(registerScript.Do(
		conn,
		"users", // KEYS[1]
		fmt.Sprintf("user:%d", userId), // KEYS[2]
		"auths",         // KEYS[3]
		"users_by_time", // KEYS[4]
		userId,          // ARGV[1]
		username,        // ARGV[2]
		hashedPw,        // ARGV[3]
		//	auth,               // ARGV[4]
		time.Now().Unix())) // ARGV[5]
	return auth, err
}

func login(username, password string) (err error, userid string) {
	logger.Printf("redisloginprovider:login")
	userId, err := redis.Int(conn.Do("HGET", "users", username))
	if err != nil {
		logger.Printf("Error from redis: %v", err)
		return errors.New("Wrong username or password"), ""
	}
	realPassword, err := redis.String(conn.Do("HGET", fmt.Sprintf("user:%d", userId), "password"))
	if err != nil {
		return err, ""
	}

	hashedPassword := hashPw(password)

	if hashedPassword != realPassword {
		return errors.New("Wrong username or password"), ""
	} else {
		log.Println("Password correct")
	}
	return nil, strconv.Itoa(userId)
}

func hashPw(password string) string {
	hasher := hmac.New(sha256.New, key)
	hasher.Write([]byte(password))
	return string(hasher.Sum(nil))
}

func getRoleListForUser(userid string) ([]int, error) {
	v, err := redis.Strings(conn.Do("SMEMBERS", "userroles:"+userid))
	if err != nil {
		return nil, err
	}

	var values []int
	for _, e := range v {
		val, _ := strconv.Atoi(e)
		values = append(values, val)
	}
	return values, nil
}
