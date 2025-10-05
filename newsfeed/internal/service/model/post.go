package model

// TODO: more fields
type Post struct {
	ID     int64
	UserID int64

	Content          string
	CreatedTimestamp int64
}
