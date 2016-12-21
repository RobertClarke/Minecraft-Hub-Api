package mysqlauth

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func login(un string, pw string) (funcerr error, userid string) {
	//TODO: replace this with proper string insertion
	rows, err := getRows("SELECT id, username, password FROM users where username='" + un + "'")

	if err != nil {
		return err, ""
	}

	if rows != nil {
		defer rows.Close()
	}
	var (
		id       int
		name     string
		password string
	)

	var count int

	for rows.Next() {
		count++
		err = rows.Scan(&id, &name, &password)
		fmt.Printf("id: %v name:%v password:%v\n", id, name, password)
	}

	if count == 0 {
		return errors.New("unable to login"), ""
	}

	var hashed []byte
	var entered []byte

	hashed = []byte(password)
	entered = []byte(pw)

	err = bcrypt.CompareHashAndPassword(hashed, entered)

	if err != nil {
		return err, ""
	}

	return nil, strconv.Itoa(id)
}

func dbPing() {
	//prod
	un := "clarkezone"
	rows, err := getRows("SELECT id, display_name, password FROM users where username='" + un + "'")
	if rows != nil {
		defer rows.Close()
	}
	if err == nil {
		fmt.Printf("DBPring: no error\n")
	} else {
		fmt.Printf("error:%v", err.Error())
	}
}

func getRoleListForUser(userid string) ([]int, error) {
	var err error
	rows, _ := getRows("select role from users where id = '" + userid + "'")
	defer rows.Close()

	roles := make([]int, 0)

	var level int
	//TODO: this is wrong; only one user
	for rows.Next() {
		err = rows.Scan(&level)
		fmt.Printf("level: %v", level)
	}

	if level == 2 {
		roles = append(roles, 1)
	}

	return roles, err
}

func getRows(sqlQuery string) (*sql.Rows, error) {
	var err error
	var db *sql.DB

	db, err = sql.Open("mysql", `mchubapp:MDcwJjPlXXVWLoY0@tcp(45.59.121.18:3306)/mchub`)
	defer db.Close()

	if err != nil {
		fmt.Printf("unable to connect to database\n")
		return nil, err
	}

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
