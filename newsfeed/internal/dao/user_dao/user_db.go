package user_dao

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type (
	UserDAI struct {
		db *gorm.DB
	}

	UserDbConfig struct {
		Username     string
		Password     string
		Host         string
		Port         int
		DatabaseName string
	}
)

func New(conf *UserDbConfig) (*UserDAI, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.DatabaseName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %s", err)
	}

	return &UserDAI{
		db: db,
	}, nil
}

func NewWithGormDB(db *gorm.DB) (*UserDAI, error) {
	return &UserDAI{db: db}, nil
}

func (d *UserDAI) Stop() error {
	sqlDb, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDb.Close()
}

func (d *UserDAI) Create(ctx context.Context, user *model.User) (*model.User, error) {
	dbUser := &UserDbModel{
		Username:     user.Username,
		HashPassword: user.HashedPassword,
		Email:        user.Email,
		DisplayName:  user.DisplayName,
		Dob:          user.Dob,
		Removed:      false,
	}

	result := d.db.WithContext(ctx).Create(dbUser)
	if err := result.Error; err != nil {
		return nil, err
	}

	user.ID = dbUser.ID
	return toUserModel(dbUser, true), nil
}

func (d *UserDAI) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	dbUser := &UserDbModel{}
	err := d.db.WithContext(ctx).Where("user_name=? and removed=false", username).First(dbUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUserModel(dbUser, true), nil
}

func (d *UserDAI) GetByID(ctx context.Context, userId int64) (*model.User, error) {
	dbUser := &UserDbModel{}
	err := d.db.WithContext(ctx).Where("id=? and removed=false", userId).First(dbUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toUserModel(dbUser, false), nil
}

func (d *UserDAI) Follow(ctx context.Context, userId int64, peerId int64) (*model.Follow, error) {
	dbUserUser := &UserUserDbModel{}

	err := d.db.WithContext(ctx).Where("follower_id=? and following_id=?", userId, peerId).First(dbUserUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// if not found, create
	isNotFound := err != nil && errors.Is(err, gorm.ErrRecordNotFound)
	isUnfollowed := dbUserUser.Removed
	if isNotFound {
		dbUserUser.FollowerID = userId
		dbUserUser.FollowingID = peerId
		dbUserUser.FollowTimestamp = time.Now().Unix()
		dbUserUser.Removed = false

		result := d.db.WithContext(ctx).Create(dbUserUser)
		if err := result.Error; err != nil {
			return nil, err
		}
		return toFollowModel(dbUserUser, nil, nil), nil
	}

	if isUnfollowed {
		// reset follow timestamp and removed
		dbUserUser.FollowTimestamp = time.Now().Unix()
		dbUserUser.Removed = false

		result := d.db.WithContext(ctx).Save(dbUserUser)
		if err := result.Error; err != nil {
			return nil, err
		}
	}

	return toFollowModel(dbUserUser, &UserDbModel{ID: userId}, &UserDbModel{ID: peerId}), nil
}

func (d *UserDAI) Unfollow(ctx context.Context, userId int64, followingId int64) error {
	dbUserUser := &UserUserDbModel{}

	result := d.db.WithContext(ctx).Model(dbUserUser).
		Where("follower_id=? and following_id=?", userId, followingId).
		Update("removed", true)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *UserDAI) GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error) {
	lastTs, ok := paging.LastValue.(int64)
	if !ok {
		return nil, errors.New("invalid last value")
	}

	userUsers := make([]*UserUserDbModel, 0, paging.Limit)
	result := d.db.WithContext(ctx).Model(&UserUserDbModel{}).
		Select("id, following_id, follow_timestamp").
		Where("follower_id = ? AND removed = ? AND follow_timestamp < ?", userId, false, lastTs).
		Order("follow_timestamp DESC"). // sort by follow_ts desc
		Limit(int(paging.Limit)).
		Find(&userUsers)
	if result.Error != nil {
		return nil, result.Error
	}

	logger.Debug("query user_users", logger.F("user_users", userUsers))

	followings := make([]*UserDbModel, len(userUsers))
	followingIDs := make([]int64, len(userUsers))
	for i := range userUsers {
		followingIDs[i] = userUsers[i].FollowingID
	}
	err := d.db.WithContext(ctx).Where("id IN ?", followingIDs).
		Select("id, user_name, display_name, email, dob").
		Where("removed = ?", false).
		Find(&followings).Error

	logger.Debug("query users", logger.F("followings", followings))

	if err != nil {
		return nil, err
	}

	// join data wiht follows and re-sort by the correct order
	followModels := joinFollowings(userUsers, followings)
	return followModels, nil
}

func (d *UserDAI) GetFollowers(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error) {
	// TODO: implement me
	return nil, common.NewError(common.CodeNotImplemented, "DAI GetFollowers not implemented")
}

func toUserModel(user *UserDbModel, withHashedPassword bool) *model.User {
	res := &model.User{
		ID:             user.ID,
		Username:       user.Username,
		Password:       "",
		HashedPassword: "",
		DisplayName:    user.DisplayName,
		Email:          user.Email,
		Dob:            user.Dob,
	}
	if withHashedPassword {
		res.HashedPassword = user.HashPassword
	}
	return res
}

func toFollowModel(userUser *UserUserDbModel, follower *UserDbModel, following *UserDbModel) *model.Follow {
	follow := &model.Follow{
		ID:       userUser.ID,
		FollowTs: userUser.FollowTimestamp,
	}
	if follower != nil {
		follow.Follower = toUserModel(follower, false)
	}
	if following != nil {
		follow.Following = toUserModel(following, false)
	}

	return follow
}

func joinFollowings(userUsers []*UserUserDbModel, followings []*UserDbModel) []*model.Follow {
	userByIdMap := make(map[int64]*UserDbModel)
	for _, following := range followings {
		userByIdMap[following.ID] = following
	}

	followModels := make([]*model.Follow, len(followings))
	for i, userUser := range userUsers {
		followModels[i] = toFollowModel(userUser, nil, userByIdMap[userUser.FollowingID])
	}

	return followModels
}
