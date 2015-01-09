package main

import (
	"gopkg.in/mgo.v2/bson"
	"upper.io/db"
)

type User struct {
	ID     bson.ObjectId `bson:"_id,omitempty" bondb:",pk"`
	Name   string        `bson:"name"`
	Social []Social      `bson:"social"` // array of embedded docs
}

type Social struct {
	Network    string `bson:"network"`
	Username   string `bson:"username"`
	WebsiteURL string `bson:"website_url"`
}

func NewUser() *User {
	return &User{}
}

func FindUser(id bson.ObjectId) (user *User, err error) {
	err = DB.Query(&user).ID(id)
	return
}

func (u *User) CollectionName() string {
	return "users"
}

func (u *User) FindPhotos() (photos []*Photo, err error) {
	err = DB.Query(&photos).Where(db.Cond{"user_id": u.ID}).All()
	return
}

type Photo struct {
	ID     bson.ObjectId `bson:"_id,omitempty" bondb:",pk"`
	UserID bson.ObjectId `bson:"user_id,omitempty"`
	Title  string        `db:"title"`
	URL    string        `db:"url"`
}
