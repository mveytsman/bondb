package bondb

import (
	"reflect"

	"upper.io/db"
)

type Session struct {
	db.Database
}

func NewSession(adapter string, url db.ConnectionURL) (*Session, error) {
	d, err := db.Open(adapter, url)
	if err != nil {
		return nil, err
	}
	session := &Session{Database: d}
	return session, nil
}

func (s *Session) Find(cond db.Cond) Query {
	q := Query{session: s, cond: cond}
	return q
}

func (s *Session) Save(item interface{}) (interface{}, error) {
	var m CanCollectionName
	if t, ok := item.(CanCollectionName); !ok {
		return nil, ErrNonModel
	} else {
		m = t
	}

	col, err := s.Collection(m.CollectionName())
	if err != nil && err != db.ErrCollectionDoesNotExist {
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

	itemv := reflect.ValueOf(item)
	itemp := reflect.Indirect(itemv) // TODO: ... indirect always...........?

	sinfo, err := getStructInfo(itemp.Type())
	if err != nil {
		return nil, err
	}
	if sinfo != nil {
		for _, si := range sinfo.FieldsList {
			if si.PK {
				_, setter1 := item.(db.IDSetter)
				_, setter2 := item.(db.Int64IDSetter)
				_, setter3 := item.(db.Uint64IDSetter)
				if !(setter1 || setter2 || setter3) {
					itemp.FieldByName(si.Key).Set(reflect.ValueOf(oid))
				}
				break
			}
		}
	}

	if i, ok := item.(CanAfterSave); ok {
		i.AfterSave()
	}

	return oid, nil
}

// TODO or not todo ..? that is the question
// func (s *Session) Create(item interface{}) error {
// 	return nil
// }

// func (s *Session) Update(item interface{}) error {
// 	return nil
// }

// TODO
func (s *Session) Delete(item interface{}) error {
	return nil
}
