package model

type Paging struct {
	LastValue any
	Limit     int64

	// optional
	OrderBy   string
	Ascending bool
}
