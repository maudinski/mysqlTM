package main

import (
	"log"
)

func main() {
	tm, err := NewTM("root", "test", "", "testerr", "users")	
	if err != nil {  log.Fatal(err)  }

	tm.SetUnique("username", "password")
	log.Println(tm.GetByUnique("fuck"))
}
