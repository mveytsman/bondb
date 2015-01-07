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
	return &Account{}
}

func FindAccount() (account *Account, err error) {
	err = DB.Q(&account).Where(db.Cond{"name": "sup"}).One()
	return
}

func (a *Account) CollectionName() string {
	return `accounts`
}

func (a *Account) BeforeSave() error {
	// validations can go here..
	return nil
}

func (a *Account) FindUser() (user User, err error) {
	err = DB.Query(&user).Where(db.Cond{"account_id": a.Id}).One()
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

type User struct {
	Id bson.ObjectId `bson:"_id,omitempty"`
	// Id       bson.ObjectId `bson:"_id,omitempty" bondb:",pk"`
	Username string `db:"username"`

	AccountId bson.ObjectId `bson:"account_id,omitempty"`
}

func NewUser() User {
	return User{}
}

func (a User) CollectionName() string {
	return `users`
}

//--

// TODO: Mock the database. Currently a mongo database needs to be running
// in the background to run these tests.

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
	account.Disabled = true
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

	// Read
	account = nil
	err = DB.Query(&account).Where(db.Cond{"name": "Joe"}).One()

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

func TestSaveObject(t *testing.T) {
	assert := assert.New(t)
	account := &Account{Name: "Object"}
	oid, err := DB.Save(account)
	assert.NoError(err, "Saved the object")
	assert.Equal(account.Id, oid)
}

func TestSaveValue(t *testing.T) {
	assert := assert.New(t)
	user := User{Username: "Value"}
	oid, err := DB.Save(user)
	assert.NoError(err, "Saved the value")
	assert.NotEmpty(oid, "Object id has been returned")
}

func TestCreate(t *testing.T) {
	assert := assert.New(t)
	account := &Account{Name: "Object2"}
	oid, err := DB.Create(account)
	assert.NoError(err, "Saved the object")
	assert.NotEmpty(oid)
	assert.NotEqual(account.Id, oid)
	assert.Len(account.Id, 0)
}

func TestReadOne(t *testing.T) {
	assert := assert.New(t)
	var account *Account
	err := DB.Query(&account).Where(db.Cond{"name": "Joe", "disabled": true}).One()
	assert.NoError(err, "Read account record")
	assert.NotNil(account, "Found account record")
	assert.Equal("Joe", account.Name, "Account named Joe")
}

func TestReadFirst(t *testing.T) {
	assert := assert.New(t)
	var account *Account
	err := DB.Query(&account).First()
	assert.NoError(err, "Read the first record")
	assert.NotNil(account, "Found first record")
	assert.Equal("Joe", account.Name)
}

func TestReadAll(t *testing.T) {
	assert := assert.New(t)
	var accounts []*Account
	err := DB.Query(&accounts).All()
	assert.NoError(err, "Read account record")
	assert.NotEmpty(accounts, "Accounts are not empty..")
	assert.True(len(accounts) >= 2, "Got two or more accounts")
	assert.Equal("Joe", accounts[0].Name)
	assert.Equal("Peter", accounts[1].Name)
}

func TestHasOne(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	DB.Query(&account).Where(db.Cond{"name": "Joe"}).One()
	assert.Equal("Joe", account.Name, "Account named Joe")

	user, err := account.FindUser()
	assert.NoError(err, "Found user without error")
	assert.NotNil(user, "Get user object from the account")
}

func TestNotFound(t *testing.T) {
	assert := assert.New(t)

	var accounts []*Account
	err := DB.Query(&accounts).Where(db.Cond{"name": "blahblah"}).All()
	assert.NoError(err, "Unable to find object, doesn't cause an error")
	assert.Empty(accounts, "Found no accounts under that condition")

	var account *Account
	err = DB.Query(&account).Where(db.Cond{"name": "blahblah"}).One()
	assert.Error(err, "Looking for a specific object, with not found, does error")
	assert.Equal(err, db.ErrNoMoreRows)
	assert.Nil(account, "Found no account under that condition")
}

func TestSaveExistingItem(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	err := DB.Query(&account).Where(db.Cond{"name": "Peter"}).One()
	assert.NoError(err)
	pk := account.Id

	account.Name = "Piotr"
	oid, err := DB.Save(account)
	assert.NoError(err)
	assert.Equal(oid, pk, "primary key should not have changed on update")

	var account2 *Account
	err = DB.Query(&account).Where(db.Cond{"name": "Piotr"}).One()
	assert.NoError(err)
	assert.Equal(account2.Name, "Piotr")
}

// TODO........
func SkipTestUpdate(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	err := DB.Query(&account).Where(db.Cond{"name": "Joe"}).One()
	assert.NoError(err)

	account.Disabled = false
	err = DB.Update(account) //, "disabled")
	assert.NoError(err)

	var account2 *Account
	err = DB.Query(&account2).Where(db.Cond{"name": "Joe"}).One()
	assert.NoError(err)
	assert.Equal(account2.Disabled, false)
}

func TestDelete(t *testing.T) {
}
