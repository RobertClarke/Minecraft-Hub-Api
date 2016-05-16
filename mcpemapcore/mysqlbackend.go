// mysqlbackend.go
package mcpemapcore

import "fmt"

type MySqlBackend struct {
}

func (r MySqlBackend) UpdateMap(user *User,
	mapid int,
	uploadFilename string,
	pureHash string,
) {
	fmt.Printf("mapid:%v uploadFilename:%v pureHash:%v\n", mapid, uploadFilename, pureHash)
	MySqlUpdateMapNewUpload(user, mapid, pureHash)
}

func (r MySqlBackend) LoadUserInfo(userid string) (*User, error) {

	fmt.Printf("userid:%v\n", userid)
	return MySqlGetUserInfo(userid)
}
