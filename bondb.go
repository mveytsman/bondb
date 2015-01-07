package bondb

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type CanCollectionName interface {
	CollectionName() string
}

type CanBeforeSave interface {
	BeforeSave() error
}

type CanAfterSave interface {
	AfterSave()
}

type CanBeforeDelete interface {
	BeforeDelete() error
}

type CanAfterDelete interface {
	AfterDelete()
}

type CanAfterFind interface {
	AfterFind()
}

// NOTE: struct tag code borrowed + inspired from https://labix.org/mgo library
type structInfo struct {
	FieldsMap  map[string]fieldInfo
	FieldsList []fieldInfo
	Zero       reflect.Value
}

type fieldInfo struct {
	Num      int
	Key      string
	PK       bool // primary key
	Required bool // field is required
}

var structMap = make(map[reflect.Type]*structInfo)
var structMapMutex sync.RWMutex

func getStructInfo(st reflect.Type) (*structInfo, error) {
	structMapMutex.RLock()
	sinfo, found := structMap[st]
	structMapMutex.RUnlock()
	if found {
		return sinfo, nil
	}
	n := st.NumField()
	fieldsMap := make(map[string]fieldInfo)
	fieldsList := make([]fieldInfo, 0, n)
	for i := 0; i != n; i++ {
		field := st.Field(i)
		if field.PkgPath != "" {
			continue // Private field
		}

		info := fieldInfo{Num: i, Key: field.Name}
		tag := field.Tag.Get("bondb")
		if tag == "" && strings.Index(string(field.Tag), ":") < 0 {
			tag = string(field.Tag)
		}
		if tag == "-" {
			continue
		}

		attrs := strings.Split(tag, ",")
		if len(attrs) > 1 {
			for _, flag := range attrs[1:] {
				switch flag {
				case "pk":
					info.PK = true
				case "required":
					info.Required = true
				default:
					panic(fmt.Sprintf("Unsupported flag %q in tag %q of type %s", flag, tag, st))
				}
			}
			tag = attrs[0]
		}

		fieldsList = append(fieldsList, info)
		fieldsMap[info.Key] = info
	}

	sinfo = &structInfo{
		fieldsMap,
		fieldsList,
		reflect.New(st).Elem(),
	}
	structMapMutex.Lock()
	structMap[st] = sinfo
	structMapMutex.Unlock()
	return sinfo, nil
}
