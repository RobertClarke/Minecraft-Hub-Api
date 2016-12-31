package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbConStr string
)

func init() {
	//Production:
	//dbConStr = `mchubapp:MDcwJjPlXXVWLoY0@tcp(45.59.121.18:3306)/mchub`

	//dev
	dbConStr = `mchubapp:lMEsOgmYkKoo9a5v@tcp(45.59.121.15:3306)/mchub`
}

func getRowsParam(sqlQuery string, args ...interface{}) (*sql.Rows, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", dbConStr)
	//db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)
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

func getDBConnection() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", dbConStr)
	//db, err = sql.Open("mysql", `clarkezone:winBlue.,.,.,@tcp(45.59.121.13:3306)/minecrafthub_dev2?parseTime=true`)

	if err != nil {
		return nil, err
	}
	return db, nil
}

func getRowsParamFromConnection(db *sql.DB, sqlQuery string, args ...interface{}) (*sql.Rows, error) {
	var err error
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

func getResultParamFromConnection(db *sql.DB, sqlQuery string, args ...interface{}) (sql.Result, error) {
	var err error
	stmt, err := db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(args...)

	return result, err
}

func executeGetRowsParamFromConnection(stmt *sql.Stmt, args ...interface{}) (*sql.Rows, error) {
	var err error

	rows, err := stmt.Query(args...)

	if err != nil {
		return nil, err
	}

	return rows, nil
}

func prepareGetRowsParamFromConnection(db *sql.DB, sqlQuery string, args ...interface{}) (*sql.Stmt, error) {
	var err error
	stmt, err := db.Prepare(sqlQuery)
	if err != nil {
		return nil, err
	}
	return stmt, nil
}
