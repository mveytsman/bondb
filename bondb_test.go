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
	Id       bson.ObjectId `bson:"_id,omitempty"`
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

	// TODO:
	// return DB.Update(a, "disabled")

	// NOTE: updating is quite verbose
	q := DB.Query(&a).Where(db.Cond{"_id": a.Id})
	return q.Update()
}

type User struct {
	Id       bson.ObjectId `bson:"_id,omitempty"`
	Username string        `db:"username"`

	AccountId bson.ObjectId `bson:"account_id,omitempty"`
}

func NewUser() User {
	return User{}
}

func (a User) CollectionName() string {
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
	if DB == nil {
		return false
	}
	err := DB.Ping()
	if err != nil {
		return false
	}
	return true
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

func TestIntegration(t *testing.T) {
	assert := assert.New(t)

	// Create
	account := NewAccount()
	account.Name = "Joe"
	account.Disabled = true
	oid, err := DB.Create(account)
	assert.NoError(err, "Create new account record")
	assert.NotNil(oid, "Create returns the new primary key")
	// assert.Equal(account.Id, oid, "Automatically sets the primary key on the account")

	user := NewUser()
	user.Username = "joepro"
	user.AccountId = oid.(bson.ObjectId) // from above
	oid, err = DB.Create(user)
	assert.NoError(err, "Create new user record")
	assert.NotNil(oid, "Create returns the new primary key")

	account2 := NewAccount()
	account2.Name = "Peter"
	oid, err = DB.Create(account2)
	account2.Id = oid.(bson.ObjectId)
	assert.NoError(err, "Create new account record")
	assert.NotNil(oid, "Create returns the new primary key")

	// Read
	account = nil
	err = DB.Query(&account).Where(db.Cond{"name": "Joe"}).One()
	assert.NoError(err, "Read account record")
	assert.NotNil(account, "Found account record")
	assert.Equal("Joe", account.Name, "Account named Joe")

	// Update
	q := DB.Query(&account2).Where(db.Cond{"_id": account2.Id})
	account2.Name = "Pete"
	err = q.Update()
	assert.NoError(err)
	var accountChk *Account
	err = DB.Query(&accountChk).Where(db.Cond{"_id": account2.Id}).One()
	assert.Equal(account2.Name, accountChk.Name)

	// Delete
	account3 := &Account{Name: "Mitch"}
	oid, err = DB.Create(account3)
	assert.NoError(err)
	account3.Id = oid.(bson.ObjectId)
	err = DB.Query(&account3).Where(db.Cond{"_id": account3.Id}).Remove()
	assert.NoError(err)
}

func TestCreateObject(t *testing.T) {
	assert := assert.New(t)
	account := &Account{Name: "Object"}
	oid, err := DB.Create(account)
	assert.NoError(err, "Saved the object")
	assert.NotEmpty(oid)
	assert.NotEqual(account.Id, oid)
	assert.Len(account.Id, 0)
}

func TestCreateValue(t *testing.T) {
	assert := assert.New(t)
	user := User{Username: "Value"}
	oid, err := DB.Create(user)
	assert.NoError(err, "Saved the value")
	assert.NotEmpty(oid, "Object id has been returned")
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
	assert.Equal("Pete", accounts[1].Name)
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

func TestUpdateItemAfterQuery(t *testing.T) {
	assert := assert.New(t)

	var account *Account
	q := DB.Query(&account).Where(db.Cond{"name": "Pete"})
	err := q.One()
	assert.NoError(err)

	account.Name = "Piotr"
	err = q.Update()
	assert.NoError(err)

	var account2 *Account
	err = DB.Query(&account2).Where(db.Cond{"name": "Piotr"}).One()
	assert.NoError(err)
	assert.Equal(account2.Name, "Piotr")
}

func TestDelete(t *testing.T) {
}
