package bondb

import (
	"fmt"
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
	col, err := s.GetCollection(item)
	if err != nil {
		return nil, err
	}
	if i, ok := item.(CanBeforeSave); ok {
		err := i.BeforeSave()
		if err != nil {
			return nil, err
		}
	}
	oid, err := col.Append(item)
	if err != nil {
		return nil, err
	}
	if i, ok := item.(CanAfterSave); ok {
		i.AfterSave()
	}
	return oid, nil
}

func (s *Session) Save(item interface{}) error {
	col, err := s.GetCollection(item)
	if err != nil {
		return err
	}

	itemv := reflect.ValueOf(item)
	oid, idkey, err := s.getPrimaryKey(itemv)
	if err != nil {
		return err
	}
	if idkey == "" {
		panic("Save() expects a struct with a 'pk' tag defined")
	}

	if i, ok := item.(CanBeforeSave); ok {
		err := i.BeforeSave()
		if err != nil {
			return err
		}
	}
	if oid == nil {
		// New
		oid, err = col.Append(item)
		if err != nil {
			return err
		}
		err = s.setPrimaryKey(itemv, oid)
		if err != nil {
			return err
		}
	} else {
		// Existing
		err := col.Find(db.Cond{idkey: oid}).Update(item)
		if err != nil {
			return err
		}
	}
	if i, ok := item.(CanAfterSave); ok {
		i.AfterSave()
	}
	return nil
}

func (s *Session) Delete(item interface{}) error {
	col, err := s.GetCollection(item)
	if err != nil {
		return err
	}
	if i, ok := item.(CanBeforeDelete); ok {
		err := i.BeforeDelete()
		if err != nil {
			return err
		}
	}
	itemv := reflect.ValueOf(item)
	oid, idkey, err := s.getPrimaryKey(itemv)
	if err != nil {
		return err
	}
	err = col.Find(db.Cond{idkey: oid}).Remove()
	if err != nil {
		return err
	}
	if i, ok := item.(CanAfterDelete); ok {
		i.AfterDelete()
	}
	return nil
}

func (s *Session) GetCollection(item interface{}) (db.Collection, error) {
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

func (s *Session) ReflectCollection(v reflect.Value) (db.Collection, error) {
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
	return s.GetCollection(item)
}

func (s *Session) getPrimaryKey(itemv reflect.Value) (interface{}, string, error) {
	if itemv.Kind() != reflect.Ptr {
		return nil, "", db.ErrExpectingPointer
	}
	itemp := reflect.Indirect(itemv)

	var i reflect.Value
	if itemp.Kind() == reflect.Struct {
		i = itemp
	} else {
		i = reflect.Indirect(itemp)
	}
	if !i.IsValid() {
		panic(fmt.Sprintf("invalid type passed: %v", itemv.Type()))
	}

	sinfo, err := getStructInfo(i.Type())
	if err != nil {
		return nil, "", err
	}
	pkInfo := sinfo.PKFieldInfo
	if pkInfo == nil {
		return nil, pkInfo.Key, nil // ...? hmm.. return error...?
	}

	pk := i.FieldByName(pkInfo.Name)
	v := pk.Interface()
	z := pkInfo.Zero.Interface()

	if v == nil || v == z {
		return nil, pkInfo.Key, nil
	}
	return v, pkInfo.Key, nil
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
			f := itemp.FieldByName(fi.Name)
			if f.CanSet() {
				f.Set(reflect.ValueOf(oid))
			} else {
				panic(fmt.Sprintf("cannot set the value of type:%s field:sq", itemp.Type(), fi.Name))
			}
		}
	}
	return nil
}
