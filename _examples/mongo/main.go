package main

import (
	"log"

	"github.com/pressly/bondb"
	"upper.io/db"
	_ "upper.io/db/mongo"
)

var (
	DB *bondb.Session
)

func main() {
	var err error
	DB, err = bondb.NewSession("mongo", db.Settings{
		Host:     "127.0.0.1",
		Database: "bonexample",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Reset the db
	cols, _ := DB.Collections()
	for _, k := range cols {
		col, err := DB.Collection(k)
		if err == nil {
			col.Truncate()
		}
	}

	//--

	u := &User{Name: "Peter", Social: []Social{
		Social{"twitter", "@peterk", "http://twitter.com/peterk"},
	}}
	err = DB.Save(u)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Saved the new user:", u)

	u.Name = "Batman"
	err = DB.Save(u)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Updated the user:", u)
}
