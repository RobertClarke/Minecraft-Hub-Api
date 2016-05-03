package mysqlauth

import "fmt"

type MysqlAuthProvider struct {
}

func (a MysqlAuthProvider) Login(username string, password string) (result bool, userid string) {
	err, id := login(username, password)
	if err != nil {
		fmt.Printf("Login:Error:%v", err.Error())
		return false, ""
	}

	fmt.Printf("Login:%v\n", username)
	return true, id
}

func (a MysqlAuthProvider) GetRoles(userid string) []int {
	roles, _ := getRoleListForUser(userid)
	return roles
}
