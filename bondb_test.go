package bondb_test

import (
	"os"
	"testing"

	"github.com/pressly/bondb"
	"upper.io/db"
	_ "upper.io/db/mongo"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

var (
	DB *bondb.Session
)

type Account struct {
	Id       bson.ObjectId `bson:"_id,omitempty" bondb:",pk,required"`
	Name     string        `db:"name"`
	Disabled bool          `db:"disabled"`
}

func NewAccount() *Account {
	a := &Account{}
	a.SetDefaults()
	return a
}

func FindAccount() (account *Account, err error) {
	err = DB.Find(db.Cond{"name": "sup"}).One(&account)
	return
}

func (a *Account) CollectionName() string {
	return `accounts`
}

func (a *Account) SetDefaults() { // .. CanSetDefaults ......? which we do before a find..?
}

func (a *Account) BeforeSave() error {
	// validations here........
	return nil
}

func (a *Account) FindUser() (user *User, err error) {
	err = DB.Find(db.Cond{"account_id": a.Id}).One(&user)
	return
}

func (a *Account) ToggleDisabled() error {
	a.Disabled = !a.Disabled

	// return DB.Update(db.Fields{"disabled"})
	// return DB.Only(bondb.Fields{"disabled"}).Save(a) // ...?
	// return DB.Save(a)

	//.. it appears that db.Result has an Update() method that
	// does exactly this..

	return nil
}

// func (a *Account) QueryImages() {
// 	var images []Image
// 	// DB.Find(bondb.Cond{"account_id": a.Id}).All(&images)
// 	DB.FindAll(db.Cond{"account_id": a.Id}, &images)
// 	a.Images = images
// }

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `db:"username"`

	AccountId bson.ObjectId `bson:"account_id,omitempty"`
}

func NewUser() *User {
	return &User{}
}

func (a *User) CollectionName() string {
	return `users`
}

//--

func init() {
	DB, _ = bondb.NewSession("mongo", db.Settings{
		Host:     "127.0.0.1",
		Database: "bondb_test",
	})
}

func dbConnected() bool {
	return DB != nil
}

func dbReset() {
	cols, _ := DB.Collections()
	for _, k := range cols {
		col, err := DB.Collection(k)
		if err == nil {
			col.Truncate()
		}
	}
}

func TestMain(t *testing.M) {
	status := 0
	if dbConnected() {
		dbReset()
		status = t.Run()
	} else {
		status = -1
	}
	os.Exit(status)
}

func TestConnection(t *testing.T) {
	assert := assert.New(t)
	err := DB.Ping()
	assert.NoError(err, "Ping the database")
}

func TestIntegration(t *testing.T) {
	assert := assert.New(t)

	// Create
	account := NewAccount()
	account.Name = "Joe"
	oid, err := DB.Save(account)
	assert.NoError(err, "Save new account record")
	assert.NotNil(oid, "Save returns the new primary key")
	assert.Equal(account.Id, oid, "Automatically sets the primary key on the account")

	user := NewUser()
	user.Username = "joepro"
	user.AccountId = oid.(bson.ObjectId) // from above
	oid, err = DB.Save(user)
	assert.NoError(err, "Save new user record")
	assert.NotNil(oid, "Save returns the new primary key")

	account2 := NewAccount()
	account2.Name = "Peter"
	oid, err = DB.Save(account2)
	assert.NoError(err, "Save new account record")
	assert.NotNil(oid, "Save returns the new primary key")

	// assert.Equal(account2.Id, oid) // TODO ... can we do this automatically with SetID() ..?

	// TODO: what happens if we do a DB.Save(account2) again now..? after its been saved..
	// currently, it will make a new record..

	// Read
	account = nil
	err = DB.Find(db.Cond{"name": "Joe"}).One(&account)

	assert.NoError(err, "Read account record")
	assert.NotNil(account, "Found account record")
	assert.Equal("Joe", account.Name, "Account named Joe")

	// Update
	// TODO
	// hmm.. we can either have
	// DB.Update() and DB.Create()
	// or we have just a single .Save() but the record will require Model struct to be embedded..
	// both options..?

	// Delete
	// TODO
}

func TestReadOne(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	err := DB.Find(db.Cond{"name": "Joe"}).One(&account)

	assert.NoError(err, "Read account record")
	assert.NotNil(account, "Found account record")
	assert.Equal("Joe", account.Name, "Account named Joe")
}

func TestReadFirst(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	err := DB.Find(db.Cond{}).First(&account)

	assert.NoError(err, "Read the first record")
	assert.NotNil(account, "Found first record")
	assert.Equal("Joe", account.Name)
}

func TestReadAll(t *testing.T) {
	assert := assert.New(t)

	var accounts []*Account
	err := DB.Find(db.Cond{}).All(&accounts)

	assert.NoError(err, "Read account record")
	assert.NotEmpty(accounts, "Accounts are not empty..")
	assert.True(len(accounts) >= 2, "Got two or more accounts")
	assert.Equal("Joe", accounts[0].Name)
	assert.Equal("Peter", accounts[1].Name)
}

func TestHasOne(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	DB.Find(db.Cond{"name": "Joe"}).One(&account)
	assert.Equal("Joe", account.Name, "Account named Joe")

	user, err := account.FindUser()
	assert.NoError(err, "Found user without error")
	assert.NotNil(user, "Get user object from the account")
}
