// mysqlbackend.go
package mcpemapcore

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type MySqlBackend struct {
}

func (r MySqlBackend) CreateMap(user *User,
	newMap *NewMap) {

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

func (r MySqlBackend) UpdateFavoriteMap(u *User, mapId string, fav bool) error {
	var err error
	db, err := sqlx.Connect("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)
	if err != nil {
		log.Fatal(err)
	}

	tx := db.MustBegin()
	db.MustExec("delete from favorites where user = ? and post = ? and type='map'", u.Id, mapId)
	if fav {
		db.MustExec("insert into favorites (user, post, type) values(?, ?, 'map')", u.Id, mapId)
	}
	db.MustExec("update content_maps set favorites = (select count(*) from favorites where post=? and type='map') where id=?", mapId, mapId)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (r MySqlBackend) UpdateMapDownloadCount(hash string) {
	var err error
	db, err := sqlx.Connect("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)
	if err != nil {
		log.Fatal(err)
	}

	tx := db.MustBegin()

	themap := Map{}
	err = db.Get(&themap, "select title, downloads from content_maps where file_hash=?", hash)
	if err != nil {
		log.Fatal(err)
	}

	db.MustExec("update content_maps set downloads=? where file_hash=?", themap.DownloadCount+1, hash)
	tx.Commit()
}

func (r MySqlBackend) createMap(hash string) {
	var err error
	db, err := sqlx.Connect("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)
	if err != nil {
		log.Fatal(err)
	}

	tx := db.MustBegin()
	panic("not implemented")
	tx.Commit()
}
