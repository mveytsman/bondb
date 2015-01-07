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


err = DB.Query(&account).Where(db.Cond{}).One()
err = DB.Query(&accounts).Where(db.Cond{}).Sort().Limit().All()
err = DB.Query(&account).Where().First()
err = DB.Query(&account).Sort("-_id").Last()

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

err = DB.Query(&account).Id("some-id")

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

# TODO:

DONE 1. Query and db.Result combination.. like DB.Query(&account).Where().One() etc. etc.

DONE 2. DB.Create() 

3. DB.Update()

4. DB.Delete()

5. DB.Query(&account).Id("some-id")

6. what happens if we do a DB.Save(account2) again now..? after its been saved.... UPDATE!!!

7. Update specific fields of an object.. not the entire object..

8. Future: required struct tag.. only basic validation support

9. Make sure we're calling callbacks.......

*. Error checking.. what if the db fails to save, or query..? ie. mysql key constraint
