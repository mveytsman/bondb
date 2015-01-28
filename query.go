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

func (q *query) ID(v interface{}) error {
	_, idkey, err := q.session.getPrimaryKey(q.dstv)
	if err != nil {
		return err
	}

	err = q.Result.Where(db.Cond{idkey: v}).One(q.dst)
	if err != nil {
		return err
	}
	q.PostFind()
	return nil
}

func (q *query) One() error {
	if q.err != nil {
		return q.err
	}
	err := q.Result.One(q.dst)
	if err != nil {
		return err
	}
	q.PostFind()
	return nil

}

func (q *query) First() error {
	if q.err != nil {
		return q.err
	}
	err := q.Result.One(q.dst)
	if err != nil {
		return err
	}
	q.PostFind()
	return nil

}

// TODO: add Last() error method

func (q *query) All() error {
	if q.err != nil {
		return q.err
	}
	if q.dstv.Elem().Kind() != reflect.Slice {
		return db.ErrExpectingSlicePointer
	}
	err := q.Result.All(q.dst)
	if err != nil {
		return err
	}
	q.PostFind()
	return nil

}

// empty fieldList updates all fields
func (q *query) Update(fieldList ...string) error {
	if len(fieldList) > 0 {
		updateMap := make(map[string]interface{})
		s := reflect.Indirect(q.dstv.Elem())
		for _, field := range fieldList {
			updateMap[field] = s.FieldByName(field).Interface()
		}
		err := q.Result.Update(updateMap)
		if err != nil {
			return err
		}
	} else {
		err := q.Result.Update(q.dst)
		if err != nil {
			return err
		}
	}
	return nil
}

func (q *query) Remove() error {
	item := q.dstv.Elem().Interface()
	if i, ok := item.(CanBeforeDelete); ok {
		err := i.BeforeDelete()
		if err != nil {
			return err
		}
	}
	err := q.Result.Remove()
	if err != nil {
		return err
	}
	if i, ok := item.(CanAfterDelete); ok {
		i.AfterDelete()
	}
	return nil
}

func (q *query) Close() error {
	return q.Result.Close()
}

//Called after a find. Converts time fields that need it to UTC and calls AfterFind if exists
func (q *query) PostFind() {
	item := q.dstv.Elem().Interface()
	if i, ok := item.(CanAfterFind); ok {
		i.AfterFind()
	}
}
