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
var structMap = make(map[reflect.Type]*structInfo)
var structMapMutex sync.RWMutex

type structInfo struct {
	FieldsList  []fieldInfo
	Zero        reflect.Value
	PKFieldInfo *fieldInfo
}

type fieldInfo struct {
	Index    int
	Name     string
	Tag      reflect.StructTag
	Zero     reflect.Value
	Key      string // db field key
	PK       bool   // primary key flag
	Required bool   // required field flag
}

func getStructInfo(st reflect.Type) (*structInfo, error) {
	structMapMutex.RLock()
	sinfo, found := structMap[st]
	structMapMutex.RUnlock()
	if found {
		return sinfo, nil
	}

	n := st.NumField()
	fieldsList := make([]fieldInfo, 0, n)
	var pkFieldInfo *fieldInfo

	for i := 0; i != n; i++ {
		field := st.Field(i)
		if field.PkgPath != "" {
			continue // Private field
		}

		info := fieldInfo{
			Index: i,
			Name:  field.Name,
			Tag:   field.Tag,
			Zero:  reflect.New(field.Type).Elem(),
		}

		info.Key = info.Tag.Get("db")
		if info.Key == "" {
			info.Key = info.Tag.Get("field")
		}
		if info.Key == "" {
			info.Key = info.Tag.Get("bson")
		}
		if info.Key != "" {
			parts := strings.Split(info.Key, ",")
			info.Key = parts[0]
		}
		if info.Key == "" || info.Key == "-" {
			continue
		}

		attrs := strings.Split(field.Tag.Get("bondb"), ",")
		if len(attrs) > 1 {
			for _, flag := range attrs[1:] {
				switch flag {
				case "pk":
					info.PK = true
					pkFieldInfo = &info
				case "required":
					info.Required = true
				default:
					panic(fmt.Sprintf("Unsupported flag %q in tag %q of type %s", flag, info.Key, st))
				}
			}
		}

		fieldsList = append(fieldsList, info)
	}

	sinfo = &structInfo{
		fieldsList,
		reflect.New(st).Elem(),
		pkFieldInfo,
	}
	structMapMutex.Lock()
	structMap[st] = sinfo
	structMapMutex.Unlock()
	return sinfo, nil
}
