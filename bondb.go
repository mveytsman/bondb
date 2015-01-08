package bondb

type CanCollectionName interface {
	CollectionName() string
}

// this can work too...........
// type Model interface {
//   CollectionName() string
//   PrimaryKey() interface{}
//   SetPrimaryKey(interface{})
// }

// or...
// type CanPrimaryKey interface {
//   PrimaryKey() interface{}
//   SetPrimaryKey(interface{})
// }

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
