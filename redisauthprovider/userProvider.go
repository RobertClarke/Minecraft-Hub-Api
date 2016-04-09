package redisauth

import (
	"fmt"
	"strconv"

	"github.com/garyburd/redigo/redis"
)

type RedisUserProvider struct {
}

func (a RedisUserProvider) Login(username string, password string) (result bool, userid string) {
	err, id := login(username, password)
	if err != nil {
		fmt.Printf("Login:Error:%v", err.Error())
		return false, ""
	}

	fmt.Printf("Login:%v\n", username)
	return true, id
}

func (a RedisUserProvider) GetRoles(userid string) []int {
	roles, _ := getRoleListForUser(userid)
	return roles
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
