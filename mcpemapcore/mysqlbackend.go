// mysqlbackend.go
package mcpemapcore

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

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

func (r MySqlBackend) UpdateFavoriteMap(u *User, mapId string, fav bool) error {
	var err error
	db, err := sqlx.Connect("mysql", `clarkezone:winblue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parsetime=true`)
	if err != nil {
		log.Fatal(err)
	}

	tx := db.MustBegin()

	db.MustExec("insert into favorites user, post, type values(?, ?, 'map')", u.Id, mapId)
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	return err
}

func (r MySqlBackend) UpdateMapDownloadCount(hash string) {
	var err error
	db, err := sqlx.Connect("mysql", `clarkezone:winblue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parsetime=true`)
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
