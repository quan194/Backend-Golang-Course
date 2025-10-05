package model

type User struct {
	ID             int64
	Username       string
	Password       string `json:"-"`
	HashedPassword string `json:"-"`
	DisplayName    string
	Email          string
	Dob            string
}

type Follow struct {
	ID        int64
	Follower  *User // optional
	Following *User // optional
	FollowTs  int64
}
