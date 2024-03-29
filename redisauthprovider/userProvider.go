package redisauth

import "fmt"

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
