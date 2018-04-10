package db

import "fmt"

var (
	// ErrMetaNotFound is returned when meta bucket hasn't been
	// created.
	ErrMetaNotFound = fmt.Errorf("unable to locate meta information")
)
