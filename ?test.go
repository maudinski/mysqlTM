package main

import (
	"log"
	"database/sql"
	_"github.com/go-sql-driver/mysql"
)

func main(){
	db, err := sql.Open("mysql", "root:test@/testerr")
	if err != nil { log.Fatal(err) }	
	
	vals := make([]interface{}, 2)
	vals[0], vals[1] = "guts", "stug"

	_, err = db.Exec("insert into users(username, screenname) values(?, ?)", vals...)
	if err != nil { log.Fatal(err) }
}
