package bondb

import (
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

	var oid interface{}

	oid, err = col.Append(item)
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
	// NOTE: requires PrimaryKey functionality..
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
