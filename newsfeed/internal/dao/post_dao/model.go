package post_dao

type PostDbModel struct {
}

func (PostDbModel) TableName() string {
	return "posts"
}
