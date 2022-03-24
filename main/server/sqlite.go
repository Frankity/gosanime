package server

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func SqlInit() {
	db, err := sql.Open("sqlite3", "./gosanime.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	creatStmt := `CREATE TABLE users (id integer not null primary key, email text)`

	_, err = db.Exec(creatStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, creatStmt)
		return
	}

	favStmt := `CREATE table favs (id integer not null primary key,
		anime text, poster text, type text, synopsis text)`

	_, err = db.Exec(favStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, favStmt)
		return
	}

}

func StoreUser(email string) bool {

	success := false

	db, err := sql.Open("sqlite3", "./gosanime.db")
	if err != nil {
		log.Fatal(err)
	}

	is, err := db.Begin()
	if err != nil {
		success = false
		log.Fatal(err)
	}

	insertStmt, err := is.Prepare(fmt.Sprintf("INSERT INTO users(email) values(%s)", email))
	if err != nil {
		success = false
		log.Fatal(err)
	} else {
		success = true
	}

	defer insertStmt.Close()
	is.Commit()

	return success
}
