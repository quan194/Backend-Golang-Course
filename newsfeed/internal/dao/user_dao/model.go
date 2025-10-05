package user_dao

type UserDbModel struct {
	ID           int64  `gorm:"column:id"`
	Username     string `gorm:"column:user_name"`
	HashPassword string `gorm:"column:hash_password"`
	Email        string `gorm:"column:email"`
	DisplayName  string `gorm:"column:display_name"`
	Dob          string `gorm:"column:dob"`
	Removed      bool   `gorm:"column:removed"`
}

func (UserDbModel) TableName() string {
	return "users"
}

type UserUserDbModel struct {
	ID              int64 `gorm:"column:id"`
	FollowerID      int64 `gorm:"column:follower_id"`
	FollowingID     int64 `gorm:"column:following_id"`
	FollowTimestamp int64 `gorm:"column:follow_timestamp"`
	Removed         bool  `gorm:"column:removed"`
}

func (UserUserDbModel) TableName() string {
	return "user_users"
}
