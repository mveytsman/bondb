Bon DB adapter
==============

Features:
* Selectors
* Before/after hooks
* Model 


```go
conf := bondb.Conf{} // ...?
db, err := bondb.NewSession(conf)

// Create new account
account := Account{Name:"sup"}
db.Save(account)
db.Create(account) // .. like this...?

// Find an account
var account Account
err := db.Find(bondb.Cond{"name":"sup"}).One(account)

// Update an account
db.Save(account)

// Delete an account
db.Delete(account)

```

~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
or...

```go






err = DB.One(&account).Where(db.Cond{}).Do()

err = DB.All(&accounts).Where(db.Cond{}).Sort().Limit().Do()

err = DB.First(&account).Where().Do()

err = DB.Last(&account).Sort("-_id").Do()



err = DB.Query(&accounts).Where(db.Cond{}).Last()
err = DB.Q(&account).First()




























err := DB.Find(&account).WhereId("some-id")































^------- but, what is the thing that actually runs the query?

or..














err = DB.Find(db.Cond{}).One(&account)
err = DB.Find(db.Cond{}).Sort().Limit().All(&accounts)





















or..

------------- pretty good
err = DB.One(&account).Where(db.Cond{}).Query()
err = DB.All(&accounts).Where(db.Cond{}).Sort().Limit().Query()
err = DB.First(&account).Where().Query()
err = DB.Last(&account).Sort("-_id").Query()
-------------- epg


err = DB.Last(&account).Sort("-_id").Do()



```


# Associations

1. One-to-One
2. One-to-Many
3. Many-to-Many

#-----------

# TODO:

1. Query and db.Result combination.. like DB.One(&account).Where() etc. etc.

2. Check struct tag properties for ,pk  .. which tells us primary key..? which we use for SetID()

3. Error checking.. what if the db fails to save, or query..? ie. mysql key constraint

4. Test Bondb with *Account and Account
