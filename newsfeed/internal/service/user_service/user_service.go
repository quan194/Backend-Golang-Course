package user_service

import (
	"context"
	"reflect"

	"golang.org/x/crypto/bcrypt"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type UserDAI interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetByID(ctx context.Context, userId int64) (*model.User, error)

	Follow(ctx context.Context, userId, peerId int64) (*model.Follow, error)
	Unfollow(ctx context.Context, userId int64, peerId int64) error

	GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error)
	GetFollowers(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error)
}

type UserCacheDAI interface {
	SetCachedUser(ctx context.Context, user *model.User) error
	GetCachedUserByID(ctx context.Context, userId int64) (*model.User, error)

	AddCachedFollow(ctx context.Context, follow *model.Follow) error
	GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error)
}

type UserService struct {
	dai UserDAI

	enabledCache bool
	cacheDai     UserCacheDAI
}

func New(userDai UserDAI, userCacheDai UserCacheDAI) (*UserService, error) {
	svc := &UserService{
		dai:      userDai,
		cacheDai: userCacheDai,
	}

	if userCacheDai == nil || reflect.ValueOf(userCacheDai).IsNil() {
		svc.enabledCache = false
	} else {
		svc.enabledCache = true
	}

	return svc, nil
}

func (s *UserService) Signup(ctx context.Context, user *model.User) (*model.User, error) {
	existedUser, err := s.dai.GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}
	if existedUser != nil {
		return nil, common.NewError(common.CodeExistedUsername, "username is existed")
	}

	hashedPassword, err := hashPassword(user.Password)
	if err != nil {
		return nil, common.WrapError(common.CodeInternal, "failed to hash password", err)
	}
	user.HashedPassword = hashedPassword

	res, err := s.dai.Create(ctx, user)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}

	if s.enabledCache {
		if err := s.cacheDai.SetCachedUser(ctx, user); err != nil {
			logger.Error("failed to set cache grpc", logger.E(err))
		}
	}

	user = res
	return user, nil
}

func (s *UserService) Login(ctx context.Context, user *model.User) (*model.User, error) {
	existedUser, err := s.dai.GetByUsername(ctx, user.Username)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}
	if existedUser == nil {
		return nil, common.NewError(common.CodeNotExistedUsername, "username is not existed")
	}

	matched := checkPassword(existedUser.HashedPassword, user.Password)
	if !matched {
		return nil, common.NewError(common.CodeInvalidLogin, "username or password is wrong")
	}

	return existedUser, nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func checkPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *UserService) Follow(ctx context.Context, userId, peerId int64) (*model.Follow, error) {
	user, err := s.getUserByIDFromCacheOrDb(ctx, userId)
	if err != nil {
		return nil, err
	}

	peer, err := s.getUserByIDFromCacheOrDb(ctx, peerId)
	if err != nil {
		return nil, err
	}

	f, err := s.follow(ctx, userId, peerId)
	if err != nil {
		return nil, err
	}

	f.Follower = user
	f.Following = peer
	return f, nil
}

func (s *UserService) Unfollow(ctx context.Context, userId, peerId int64) error {
	return common.NewError(common.CodeNotImplemented, "not implemented")
}

func (s *UserService) GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error) {
	_, err := s.getUserByIDFromCacheOrDb(ctx, userId)
	if err != nil {
		return nil, err
	}

	followings, err := s.getFollowingsFromCacheOrDb(ctx, userId, paging)
	if err != nil {
		return nil, err
	}

	return followings, nil
}

func (s *UserService) GetFollowers(ctx context.Context, userId int64, paging *model.Paging) ([]*model.User, error) {
	return nil, common.NewError(common.CodeNotImplemented, "not implemented yet")
}

func (s *UserService) getUserByIDFromCacheOrDb(ctx context.Context, userId int64) (*model.User, error) {
	var (
		user *model.User
		err  error
	)
	if s.enabledCache {
		user, err = s.cacheDai.GetCachedUserByID(ctx, userId)
		if err != nil {
			// monitor error here
		}
		if user != nil {
			return user, nil
		}
	}

	user, err = s.dai.GetByID(ctx, userId)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}
	if user == nil {
		return nil, common.NewError(common.CodeNotExistedUserID, "user_id is not existed")
	}

	return user, nil
}

func (s *UserService) follow(ctx context.Context, userId, peerId int64) (*model.Follow, error) {
	f, err := s.dai.Follow(ctx, userId, peerId)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}

	if s.enabledCache {
		if err = s.cacheDai.AddCachedFollow(ctx, f); err != nil {
			logger.Error("failed to set cache grpc", logger.E(err))
		}
	}

	return f, nil
}

func (s *UserService) getFollowingsFromCacheOrDb(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error) {
	var (
		followings []*model.Follow
		err        error
	)
	if s.enabledCache {
		followings, err = s.cacheDai.GetFollowings(ctx, userId, paging)
		if err != nil {
			logger.Error("failed to get followings from cache", logger.E(err))
		}
		if len(followings) > 0 {
			logger.Debug("get followings from cache")
			return followings, nil
		}
	}

	followings, err = s.dai.GetFollowings(ctx, userId, paging)
	if err != nil {
		return nil, common.WrapError(common.CodeDatabaseError, "database error", err)
	}
	logger.Debug("get followings from DB")
	return followings, nil
}
