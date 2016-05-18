package mcpemapcore

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func MySqlGetAllMaps(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=1
	order by edited desc
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, start, count)
}

func MySqlGetAllMapsInclUntested(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	order by edited desc
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, start, count)
}

func MySqlGetFeaturedMaps(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=1 and featured=1
	order by submitted
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, start, count)
}

func MySqlGetMostDownloadedMaps(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=1
	order by downloads desc
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, start, count)
}

func MySqlGetUserFavoriteMaps(user *User, start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=1 and id in (select post from favorites where type='map' and user=?
	order by downloads desc
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, user.Id, start, count)
}

func MySqlAdminGetBadMaps(start, count int, siteRoot string) ([]*AdminMap, error) {
	fmt.Printf("getsqlbadmaps\n")
	sqlQuery := "select " + getAdminMapFields() + `
	from content_maps
	inner join users on users.id=content_maps.editor_id
	where valid_direct_uri=0
	order by published desc
	limit ?, ?`
	return MySqlQueryAdminMaps(sqlQuery, siteRoot, start, count)
}

func MySqlAdminGetEditedMaps(start, count int, siteRoot string) ([]*AdminMap, error) {
	fmt.Printf("getsqleditedmaps\n")
	sqlQuery := "select " + getAdminMapFields() + `
	from content_maps
	inner join users on users.id=content_maps.editor_id
	order by edited desc
	limit ?, ?`
	return MySqlQueryAdminMaps(sqlQuery, siteRoot, start, count)
}

func MySqlGetUserInfo(userid string) (*User, error) {
	var err error
	fmt.Printf("mysqlgetuserinfo\n")
	sqlQuery := `select id, username 
	from users
	where id=?`
	us, err := MySqlQueryUsers(sqlQuery, userid)
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

func MySqlQueryUsers(sqlQuery string, args ...interface{}) ([]*User, error) {
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

func MySqlQueryMaps(sqlQuery string, siteRoot string, args ...interface{}) ([]*Map, error) {
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

func MySqlQueryAdminMaps(sqlQuery string, siteRoot string, args ...interface{}) ([]*AdminMap, error) {
	rows, err := getRowsParam(sqlQuery, args...)
	if err != nil {
		fmt.Printf("error: %v\n", err.Error())
		return nil, err
	}
	defer rows.Close()

	items, err := scanAdminMaps(rows, siteRoot)
	fmt.Printf("found %v maps\n", len(items))
	return items, err
}

func getRowsParam(sqlQuery string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	stmt, err := db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(args...)

	if err != nil {
		return nil, err
	}

	return rows, nil

}

func MySqlUpdateMapValid(theMap *Map, valid bool, uriHash string) {
	sqlQuery := `update content_maps
	set valid_direct_uri=?, file_hash=?
	where id=?
	`
	var validInt int
	validInt = 0
	if valid {
		validInt = 1
	}
	err := execParam(sqlQuery, validInt, uriHash, theMap.Id)
	if err != nil {
		fmt.Printf("error updating map: %v\n", err)
	}
}

func MySqlUpdateMapNewUpload(user *User, mapid int, hash string) {
	sqlQuery := `update content_maps
	set valid_direct_uri=1, file_hash=?, editor_id=?, edited=current_timestamp()
	where id=?
	`
	err := execParam(sqlQuery, hash, user.Id, mapid)
	if err != nil {
		fmt.Printf("error updating map: %v\n", err)
	}
}

func getMapFields() string {
	fields := []string{
		"id",
		"title",
		"description",
		"dl_link",
		"file_hash",
		"tested",
		"featured",
		"downloads",
		"images",
		"favorites"}
	return strings.Join(fields, ",")
}

func scanMaps(rows *sql.Rows, siteRoot string) ([]*Map, error) {
	var err error
	items := make([]*Map, 0)

	var title, description, dllink, imageList string
	var mapfilehash []byte
	var id, tested, downloads, featured, favorites int
	for rows.Next() {
		err = rows.Scan(
			&id,
			&title,
			&description,
			&dllink,
			//originaluri
			&mapfilehash,
			//author
			//authoruri
			//numviews
			&tested,
			&featured,
			&downloads,
			//favoritecount
			&imageList,
			&favorites)
		if err != nil {
			fmt.Printf("error: %v\n", err.Error())
			return nil, err
		}

		newMap := &Map{Id: strconv.Itoa(id),
			MapTitle:       title,
			Description:    description,
			MapDownloadUri: dllink,
			MapFileHash:    string(mapfilehash),
			DownloadCount:  int64(downloads),
			FavoriteCount:  int64(favorites)}

		if newMap.MapFileHash != "" {
			newMap.MapDownloadUri = fmt.Sprintf("%v/maps/%v.zip", siteRoot, string(mapfilehash))
		}

		if tested == 1 {
			newMap.Tested = true
		} else {

			newMap.Tested = false
		}

		if featured == 1 {
			newMap.Featured = true
		} else {

			newMap.Featured = false
		}

		imagenames := strings.Split(imageList, ",")

		for i := range imagenames {
			mi := &MapImage{}
			mi.MapImageUri = fmt.Sprintf("mcpehub.com/uploads/720x500/maps/%v", imagenames[i])
			fmt.Printf("%v\n", mi.MapImageUri)
			newMap.MapImageUriList = append(newMap.MapImageUriList, mi)
		}

		items = append(items, newMap)
	}

	return items, nil

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

		newMap := &User{Id: strconv.Itoa(id),
			Username: username,
		}

		items = append(items, newMap)
	}

	return items, nil

}
func getAdminMapFields() string {
	fields := []string{
		"content_maps.id as id",
		"title",
		"description",
		"dl_link",
		"file_hash",
		"editor_id",
		"username",
		"edited",
		"tested",
		"content_maps.featured as featured",
		"downloads",
		"images",
	}
	return strings.Join(fields, ",")
}

func scanAdminMaps(rows *sql.Rows, siteRoot string) ([]*AdminMap, error) {
	var err error
	items := make([]*AdminMap, 0)

	var title, description, dllink, imageList, editedByString string
	var mapfilehash []byte
	var id, tested, downloads, featured, editedbyid int
	var edited time.Time
	for rows.Next() {
		err = rows.Scan(
			&id,
			&title,
			&description,
			&dllink,
			//originaluri
			&mapfilehash,
			//author
			//authoruri
			&editedbyid,
			&editedByString,
			&edited,
			//numviews
			&tested,
			&featured,
			&downloads,
			//favoritecount
			&imageList)
		if err != nil {
			fmt.Printf("error: %v\n", err.Error())
			return nil, err
		}

		newMap := &AdminMap{Id: strconv.Itoa(id),
			MapTitle:       title,
			Description:    description,
			MapDownloadUri: dllink,
			MapFileHash:    string(mapfilehash),
			DownloadCount:  int64(downloads),
			EditedById:     editedbyid,
			EditedByString: editedByString,
			Edited:         edited}

		if newMap.MapFileHash != "" {
			newMap.MapDownloadUri = fmt.Sprintf("%v/maps/%v.zip", siteRoot, string(mapfilehash))
		}

		if tested == 1 {
			newMap.Tested = true
		} else {

			newMap.Tested = false
		}

		if featured == 1 {
			newMap.Featured = true
		} else {

			newMap.Featured = false
		}

		imagenames := strings.Split(imageList, ",")

		for i := range imagenames {
			mi := &MapImage{}
			mi.MapImageUri = fmt.Sprintf("mcpehub.com/uploads/720x500/maps/%v", imagenames[i])
			fmt.Printf("%v\n", mi.MapImageUri)
			newMap.MapImageUriList = append(newMap.MapImageUriList, mi)
		}

		items = append(items, newMap)
	}

	return items, nil

}

// mysql DAL functions

func execParam(sqlQuery string, args ...interface{}) error {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2`)
	defer db.Close()

	if err != nil {
		return err
	}

	stmt, err := db.Prepare(sqlQuery)
	if err != nil {
		return err
	}

	_, err = stmt.Exec(args...)
	return err
}

func getRows(sqlQuery string) (*sql.Rows, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2`)
	defer db.Close()

	if err != nil {
		return nil, err
	}

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
