package bondb

import "errors"

var (
	// ErrExpectingPointer = db.ErrExpectingPointer

	ErrNoCollectionName = errors.New("unknown collection name")
	ErrRecordNotFound   = errors.New("record not found")
)
