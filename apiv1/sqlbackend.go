package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MySqlBackend struct {
}

func (r MySqlBackend) GetAllMaps(start, count int64, siteRoot string) ([]*Map, int64, error) {
	log.Printf("GetAllMaps start:%v count:%v", start, count)
	sqlQuery := `SELECT id, title, downloads, likes, 
	(select filename from post_images where post_id=posts.id limit 1) as image, 
	(select meta_value from post_meta where post_id=posts.id and meta_key='download_link' and not(meta_value='') limit 1) as filename 
	FROM posts where type="map" order by submitted DESC
	limit ? offset ?`

	// SELECT id, title, downloads, likes, (select filename from post_images where post_id=posts.id limit 1) as image, (select meta_value from post_meta where post_id=posts.id and meta_key='download_link') as filename FROM `posts` where type="map" order by modified DESC limit 10 OFFSET 1
	//maps, err := mySQLQueryMapsProduction(sqlQuery, siteRoot)
	maps, err := mySQLQueryMapsProduction(sqlQuery, siteRoot, count, start)

	return maps, -1, err
}

func (r MySqlBackend) EnsureDirectDL(id int) (err error) {
	log.Printf("EnsureDirectDL for %v\n", id)
	return mySQLEnsureDirectDL(id)
}

func (r MySqlBackend) LoadUserInfo(userid string) (*User, error) {

	fmt.Printf("userid:%v\n", userid)
	return mySQLGetUserInfo(userid)
}

func mySQLGetUserInfo(userid string) (*User, error) {
	var err error
	fmt.Printf("mysqlgetuserinfo\n")
	sqlQuery := `select id, username 
	from users
	where id=?`
	us, err := mySQLQueryUsers(sqlQuery, userid)
	if err != nil {
		return nil, err
	}
	if len(us) == 1 {
		return us[0], err
	} else {
		fmt.Printf("user not found")
		return nil, err
	}
}

func mySQLQueryUsers(sqlQuery string, args ...interface{}) ([]*User, error) {
	rows, err := getRowsParam(sqlQuery, args...)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	items, err := scanUsers(rows)
	fmt.Printf("found %v users\n", len(items))
	return items, err
}

func scanUsers(rows *sql.Rows) ([]*User, error) {
	var err error
	items := make([]*User, 0)

	var username string
	var id int
	for rows.Next() {
		err = rows.Scan(
			&id,
			&username)
		if err != nil {
			fmt.Printf("error: %v\n", err.Error())
			return nil, err
		}

		newMap := &User{ID: strconv.Itoa(id),
			Username: username,
		}

		items = append(items, newMap)
	}

	return items, nil

}

func mySQLEnsureDirectDL(id int) error {
	existsSQL := "select meta_value from post_meta WHERE meta_key = 'direct_dl' and post_id=?"
	addSQL := "insert into post_meta(post_id ,meta_key, meta_value) values(?, 'DIRECT_DL', '1')"

	db, err := getDBConnection()

	if err != nil {
		return err
	}
	defer db.Close()

	result, err := getRowsParamFromConnection(db, existsSQL, id)

	defer result.Close()

	var metaValue int
	var found bool

	for result.Next() {
		result.Scan(&metaValue)
		found = true
	}

	if !found {
		log.Printf("None found\n")
		res, err := getResultParamFromConnection(db, addSQL, id)
		log.Printf("Meta row add result %v\n", res)
		if err != nil {
			return err
		}
	} else {
		log.Printf("Found metavalue %d\n", metaValue)
		//TODO: case where a row exists with value 0
		//above is unlikely as this function will likely only be used
		//when a new map is created
	}

	return nil
}

func mySQLQueryMapsProduction(sqlQuery string, siteRoot string, args ...interface{}) ([]*Map, error) {
	db, err := getDBConnection()

	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := getRowsParamFromConnection(db, sqlQuery, args...)
	if err != nil {
		log.Printf("error: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	items, err := scanMaps(rows, siteRoot)
	if err != nil {
		log.Printf("error during scan %v", err.Error())
		return nil, err
	}
	log.Printf("found %v maps\n", len(items))
	return items, err
}

func mySQLQueryMaps(sqlQuery string, siteRoot string, args ...interface{}) ([]*Map, error) {
	rows, err := getRowsParam(sqlQuery, args...)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	items, err := scanMaps(rows, siteRoot)
	fmt.Printf("found %v maps\n", len(items))
	return items, err
}

func scanMaps(rows *sql.Rows, siteRoot string) ([]*Map, error) {
	var err error
	items := make([]*Map, 0)

	var title, description string
	var primaryimage, downloadlink sql.NullString
	var id, downloads, favorites int
	//SELECT id, title, downloads, likes,
	//(select filename from post_images where post_id=posts.id limit 1) as image,
	//(select meta_value from post_meta where post_id=posts.id and meta_key='download_link') as filename
	//FROM posts where type="map" order by modified DESC
	//limit ? OFFSET ?
	for rows.Next() {
		err = rows.Scan(
			&id,
			&title,
			&downloads,
			&favorites,
			&primaryimage,
			&downloadlink)
		if err != nil {
			log.Printf("error: %v\n", err.Error())
			return nil, err
		}

		newMap := &Map{ID: strconv.Itoa(id),
			MapTitle:      title,
			Description:   description,
			DownloadCount: int64(downloads),
			FavoriteCount: int64(favorites)}

		if downloadlink.Valid {
			newMap.MapDownloadURI = downloadlink.String
		}

		if primaryimage.Valid {
			mi := &MapImage{}
			mi.MapImageURI = fmt.Sprintf("mcpehub.com/uploads/720x500/maps/%v", primaryimage.String)
			newMap.MapImageURIList = append(newMap.MapImageURIList, mi)

			items = append(items, newMap)
		}
	}

	return items, nil
}

func scanMapsProd(db *sql.DB, rows *sql.Rows, siteRoot string) ([]*Map, error) {
	var err error
	items := make([]*Map, 0)

	var title, description, dllink string
	var mapfilehash []byte
	var id, downloads, featured, favorites int

	stmt, err := prepareGetMapImageUrisForMap(db)

	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err = rows.Scan(
			&id,
			&title,
			&description,
			&downloads,
			&favorites)
		if err != nil {
			fmt.Printf("error: %v\n", err.Error())
			return nil, err
		}

		newMap := &Map{ID: strconv.Itoa(id),
			MapTitle:       title,
			Description:    description,
			MapDownloadURI: dllink,
			MapFileHash:    string(mapfilehash),
			DownloadCount:  int64(downloads),
			FavoriteCount:  int64(favorites)}

		if featured == 1 {
			newMap.Featured = true
		} else {
			newMap.Featured = false
		}

		imagenames, err := executeGetMapImageUrisForMap(stmt, id)
		//imagenames, err := executeGetRowsParamFromConnection(stmt, id)

		if err != nil {
			return nil, err
		}

		for i := range imagenames {
			mi := &MapImage{}
			mi.MapImageURI = fmt.Sprintf("mcpehub.com/uploads/720x500/maps/%v", imagenames[i])
			log.Printf("%v\n", mi.MapImageURI)
			newMap.MapImageURIList = append(newMap.MapImageURIList, mi)
		}

		items = append(items, newMap)
	}

	return items, nil
}

func getMapImageUrisForMap(db *sql.DB, mapid int) ([]string, error) {
	var result []string
	sqlQuery := "SELECT filename FROM `post_images` WHERE `post_id`=?"
	rows, err := getRowsParamFromConnection(db, sqlQuery, mapid)
	if err != nil {
		return nil, err
	}

	var uri string

	for rows.Next() {
		err := rows.Scan(&uri)

		if err != nil {
			return nil, err
		}
		result = append(result, uri)
	}
	return result, nil
}

func executeGetMapImageUrisForMap(stmt *sql.Stmt, mapid int) ([]string, error) {
	var result []string
	rows, err := executeGetRowsParamFromConnection(stmt, mapid)
	if err != nil {
		return nil, err
	}

	var uri string

	for rows.Next() {
		err := rows.Scan(&uri)

		if err != nil {
			return nil, err
		}
		result = append(result, uri)
	}
	return result, nil
}

func prepareGetMapImageUrisForMap(db *sql.DB) (*sql.Stmt, error) {
	sqlQuery := "SELECT filename FROM `post_images` WHERE `post_id`=?"
	statmt, err := prepareGetRowsParamFromConnection(db, sqlQuery)
	if err != nil {
		return nil, err
	}
	return statmt, nil
}

func getMapFields() string {
	fields := []string{
		"id",
		"title",
		"content",
		//"dl_link",
		//"file_hash",
		//"tested",
		//"featured",
		"downloads",
		//"images",
		"likes"}
	return strings.Join(fields, ",")
}
