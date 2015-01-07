package bondb

import (
	"log"
	"reflect"
	"sync"

	"upper.io/db"
)

type Session struct {
	db.Database

	collections     map[string]db.Collection
	collectionsLock sync.Mutex
}

func NewSession(adapter string, url db.ConnectionURL) (*Session, error) {
	d, err := db.Open(adapter, url)
	if err != nil {
		return nil, err
	}
	session := &Session{Database: d, collections: make(map[string]db.Collection)}
	return session, nil
}

func (s *Session) Query(dst interface{}) *query {
	return NewQuery(s, dst)
}

// Short-hand for Query()
func (s *Session) Q(dst interface{}) *query {
	return s.Query(dst)
}

func (s *Session) Create(item interface{}) (interface{}, error) {
	col, err := s.getCollection(item)
	if err != nil {
		return nil, err
	}

	// TODO: make sure to call callbacks

	return col.Append(item)
}

// can we specify the specific fields........? like
// bondb.Fields{"a", "b", "c"}
// or like: DB.Update(account, "name", "created_at")
// func (s *Session) Update(item interface{}, fields ...string) error {

func (s *Session) Update(item interface{}) error {
	col, err := s.getCollection(item)
	if err != nil {
		return err
	}
	_ = col

	// TODO: only allow this to work if we have a primary key..
	// so we know the _id ...
	// q := s.Query(item).Id(s.getPrimaryKey(reflect.ValueOf(item)))

	// TODO: make sure to call callbacks.....

	// return q.Update(item)
	return nil
}

// NOTE: a pk must be set on an item's struct tags in order to use this method
// ^------------- TODO
// otherwise, just use Create() and Update()
func (s *Session) Save(item interface{}) (interface{}, error) {
	// TODO: save needs a pointer..
	itemv := reflect.ValueOf(item)

	col, err := s.getCollection(item)
	if err != nil {
		return nil, err
	}

	if i, ok := item.(CanBeforeSave); ok {
		err := i.BeforeSave()
		if err != nil {
			return nil, err
		}
	}

	var oid interface{}

	pk, err := s.getPrimaryKey(itemv)
	if pk == nil {
		// New
		oid, err = col.Append(item)
		if err != nil {
			return nil, err
		}
	} else {
		// TODO: .. callbacks......... will need reorg..

		// Existing
		s.Update(item)
	}

	err = s.setPrimaryKey(itemv, oid)
	if err != nil {
		return nil, err
	}

	if i, ok := item.(CanAfterSave); ok {
		i.AfterSave()
	}

	return oid, nil
}

// TODO: delete one or a slice of objects..
func (s *Session) Delete(item interface{}) error {
	return nil
}

func (s *Session) getPrimaryKey(itemv reflect.Value) (interface{}, error) {
	// .. check isNil...  check isZero...?

	if itemv.Kind() != reflect.Ptr {
		return nil, db.ErrExpectingPointer
	}
	// if itemv.IsNil() {
	// 	return nil, nil
	// }
	itemp := reflect.Indirect(itemv)
	sinfo, err := getStructInfo(itemp.Type())
	if err != nil {
		return nil, err
	}
	log.Println("PRIMARY FIELD...====>", sinfo.PKFieldInfo)

	return nil, nil
}

func (s *Session) setPrimaryKey(itemv reflect.Value, oid interface{}) error {
	if itemv.Kind() != reflect.Ptr {
		return nil // skip, we need a pointer
	}
	itemp := reflect.Indirect(itemv)
	sinfo, err := getStructInfo(itemp.Type())
	if err != nil {
		return err
	}

	if sinfo.PKFieldInfo != nil {
		fi := sinfo.PKFieldInfo
		item := itemp.Interface()
		_, setter1 := item.(db.IDSetter)
		_, setter2 := item.(db.Int64IDSetter)
		_, setter3 := item.(db.Uint64IDSetter)
		if !(setter1 || setter2 || setter3) {
			itemp.FieldByName(fi.Key).Set(reflect.ValueOf(oid))
		}
	}

	// if sinfo != nil {
	// 	for _, si := range sinfo.FieldsList {
	// 		if si.PK {
	// 			item := itemp.Interface()
	// 			_, setter1 := item.(db.IDSetter)
	// 			_, setter2 := item.(db.Int64IDSetter)
	// 			_, setter3 := item.(db.Uint64IDSetter)
	// 			if !(setter1 || setter2 || setter3) {
	// 				itemp.FieldByName(si.Key).Set(reflect.ValueOf(oid))
	// 			}
	// 			break
	// 		}
	// 	}
	// }
	return nil
}

func (s *Session) getCollection(item interface{}) (db.Collection, error) {
	if i, ok := item.(CanCollectionName); ok {
		s.collectionsLock.Lock()
		defer s.collectionsLock.Unlock()

		colName := i.CollectionName()
		col, found := s.collections[colName]
		if found {
			return col, nil
		}
		col, err := s.Collection(colName)
		if err != nil && err != db.ErrCollectionDoesNotExist {
			return nil, err
		}

		s.collections[colName] = col
		return col, nil
	} else {
		return nil, ErrNoCollectionName
	}
}

func (s *Session) reflectCollection(v reflect.Value) (db.Collection, error) {
	var item interface{}
	if v.IsNil() || v.Kind() != reflect.Ptr {
		return nil, db.ErrExpectingPointer
	}
	if v.Elem().Kind() == reflect.Slice {
		slicev := v.Elem()
		itemT := slicev.Type().Elem()
		item = reflect.New(itemT).Elem().Interface()
	} else {
		item = reflect.Indirect(v).Interface()
	}
	return s.getCollection(item)
}
