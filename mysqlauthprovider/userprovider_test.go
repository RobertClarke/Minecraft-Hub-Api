package mysqlauth

import (
	"fmt"
	"log"
	"testing"
)

func TestDBPing(t *testing.T) {
	dbPing()
}

func TestLogin(t *testing.T) {
	var provider MysqlAuthProvider

	provider = MysqlAuthProvider{}
	res, userid := provider.Login("clarkezone", "winBlue.,.,")
	fmt.Printf("userid:%v\n", userid)
	if !res {
		t.Fail()
	}
}

func TestRoles(t *testing.T) {
	var provider MysqlAuthProvider

	provider = MysqlAuthProvider{}
	res := provider.GetRoles("1")
	if len(res) == 1 && res[0] == 1 {
		log.Println("roles correct")
	} else {
		log.Println("roles failed: len:%v role:%v", len(res), res[0])
		t.Fail()
	}
}
