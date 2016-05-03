package mysqlauth

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"strconv"
)

func login(un string, pw string) (funcerr error, userid string) {
	//TODO: replace this with proper string insertion
	rows, err := getRows("SELECT id, name, password FROM users where username='" + un + "'")
	defer rows.Close()

	var (
		id       int
		name     string
		password string
	)

	for rows.Next() {
		err = rows.Scan(&id, &name, &password)
		fmt.Printf("id: %v name:%v password:%v\n", id, name, password)
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

func getRoleListForUser(userid string) ([]int, error) {
	var err error
	rows, _ := getRows("select level from users where id = '" + userid + "'")
	defer rows.Close()

	roles := make([]int, 0)

	var level int
	//TODO: this is wrong; only one user
	for rows.Next() {
		err = rows.Scan(&level)
		fmt.Printf("level: %v", level)
	}

	if level == 9 {
		roles = append(roles, 1)
	}

	return roles, err
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
