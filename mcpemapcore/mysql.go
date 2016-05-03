package mcpemapcore

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

func MySqlGetAllMaps(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=1
	order by submitted
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

func MySqlGetBadMaps(start, count int, siteRoot string) ([]*Map, error) {
	sqlQuery := "select " + getMapFields() + `
	from content_maps
	where valid_direct_uri=0
	order by published desc
	limit ?, ?`
	return MySqlQueryMaps(sqlQuery, siteRoot, start, count)
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

func getRowsParam(sqlQuery string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2`)
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

func MySqlUpdateMap(theMap *Map, valid bool, uriHash string) {
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
		"images"}
	return strings.Join(fields, ",")
}

func scanMaps(rows *sql.Rows, siteRoot string) ([]*Map, error) {
	var err error
	items := make([]*Map, 0)

	var title, description, dllink, imageList string
	var mapfilehash []byte
	var id, tested, downloads, featured int
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
			&imageList)
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
		}

		if newMap.MapFileHash != "" {
			newMap.MapDownloadUri = fmt.Sprintf("%v/maps/%v.zip", siteRoot, mapfilehash)
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
