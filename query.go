package bondb

import (
	"reflect"

	"upper.io/db"
)

type query struct {
	session *Session
	dst     interface{}
	dstv    reflect.Value
	err     error

	Collection db.Collection
	Result     db.Result
}

func NewQuery(session *Session, dst interface{}) *query {
	q := &query{session: session, dst: dst}

	dstv := reflect.ValueOf(dst)
	if dstv.IsNil() || dstv.Kind() != reflect.Ptr {
		q.err = db.ErrExpectingPointer
		return q
	}
	q.dstv = dstv

	col, err := session.ReflectCollection(dstv)
	if err != nil {
		q.err = err
		return q
	}
	q.Collection = col
	q.Result = q.Collection.Find(db.Cond{})
	return q
}

func (q *query) Limit(v uint) *query {
	q.Result = q.Result.Limit(v)
	return q
}

func (q *query) Skip(v uint) *query {
	q.Result = q.Result.Skip(v)
	return q
}

func (q *query) Sort(v ...interface{}) *query {
	q.Result = q.Result.Sort(v...)
	return q
}

func (q *query) Select(v ...interface{}) *query {
	q.Result = q.Result.Select(v...)
	return q
}

func (q *query) Where(v ...interface{}) *query {
	q.Result = q.Result.Where(v...)
	return q
}

func (q *query) Group(v ...interface{}) *query {
	q.Result = q.Result.Group(v...)
	return q
}

func (q *query) Count() (uint64, error) {
	return q.Result.Count()
}

func (q *query) Next(v interface{}) error {
	return q.Result.Next(v)
}

func (q *query) Id(v interface{}) error {
	_, idkey, err := q.session.getPrimaryKey(q.dstv)
	if err != nil {
		return err
	}
	return q.Result.Where(db.Cond{idkey: v}).One(q.dst)
}

func (q *query) One() error {
	if q.err != nil {
		return q.err
	}
	return q.Result.One(q.dst)
}

func (q *query) First() error {
	if q.err != nil {
		return q.err
	}
	return q.Result.One(q.dst)
}

// TODO: add Last() error method

func (q *query) All() error {
	if q.err != nil {
		return q.err
	}
	if q.dstv.Elem().Kind() != reflect.Slice {
		return db.ErrExpectingSlicePointer
	}
	return q.Result.All(q.dst)
}

// TODO: take a field list.......
// func (q *query) Update(onlyFields ...string) error {

func (q *query) Update() error {
	// NOTE: Result.Update() expects a map[string]interface{} or an object with field tags
	return q.Result.Update(q.dst)
}

func (q *query) Remove() error {
	return q.Result.Remove()
}

func (q *query) Close() error {
	return q.Result.Close()
}
