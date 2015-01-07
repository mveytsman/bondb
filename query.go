package bondb

import (
	"reflect"

	"upper.io/db"
)

type Query struct {
	session *Session
	cond    db.Cond
	result  db.Result // .. hmmmmmmm..

	// .. we gotta implement the entire db.Result thing here..
}

func (q Query) One(dst interface{}) error {
	dstv := reflect.ValueOf(dst)
	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return db.ErrExpectingPointer
	}
	item := reflect.Indirect(dstv).Interface()

	col, err := q.collection(item)
	if err != nil {
		return err
	}

	// var r db.Result

	res := col.Find(q.cond)
	return res.One(dst)
}

func (q Query) First(dst interface{}) error {
	// q2 := q.Limit(1) // q2 ..?
	return q.One(dst)
}

func (q Query) Last(dst interface{}) error {
	// TODO .. like First() .. but how do we get
	// the last item from a query..? do we flip the Sort..?
	return nil
}

func (q Query) All(dst interface{}) error {
	dstv := reflect.ValueOf(dst)
	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		return db.ErrExpectingPointer
	}
	if dstv.Elem().Kind() != reflect.Slice {
		return db.ErrExpectingSlicePointer
	}

	slicev := dstv.Elem()
	itemT := slicev.Type().Elem()
	item := reflect.New(itemT).Elem().Interface()

	col, err := q.collection(item)
	if err != nil {
		return err
	}

	res := col.Find(q.cond)
	return res.All(dst)
}

func (q Query) collection(item interface{}) (db.Collection, error) {
	var m CanCollectionName
	if t, ok := item.(CanCollectionName); !ok {
		return nil, ErrNonModel
	} else {
		m = t
	}

	col, err := q.session.Collection(m.CollectionName())
	if err != nil && err != db.ErrCollectionDoesNotExist {
		return nil, err
	}
	return col, nil
}

// TODO: .. other query things.. like Where, that returns a Query ..
