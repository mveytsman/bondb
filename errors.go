package bondb

import "errors"

var (
	// ErrExpectingPointer = db.ErrExpectingPointer

	ErrNonModel       = errors.New("object is not a model")
	ErrRecordNotFound = errors.New("record not found")
)
