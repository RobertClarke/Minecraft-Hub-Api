package mcpemapcore

import (
	"strconv"

	"github.com/garyburd/redigo/redis"
)

type User struct {
	Id       string
	Username string `redis:"username"`
	Auth     string `redis:"auth"`
}

func LoadUserByUsername(username string) (*User, error) {
	userId, err := redis.Int(conn.Do("HGET", "users", username))
	if err != nil {
		return nil, err
	}
	user, err := LoadUserInfo(strconv.Itoa(userId))
	return user, err
}

func LoadUserInfo(userId string) (*User, error) {
	v, err := redis.Values(conn.Do("HGETALL", "user:"+userId))
	if err != nil {
		return nil, err
	}
	u := &User{Id: userId}
	err = redis.ScanStruct(v, u)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (ro User) AddToRole(role Role) error {
	var err error
	_, err = redis.Int(conn.Do("SADD", "userroles:"+ro.Id, role.Id))
	if err != nil {
		return err
	}
	return nil
}
