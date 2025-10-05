package user_cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type (
	CachedFollow struct {
		ID        int64
		Timestamp int64
	}
)

const (
	UserKeyFormat       = "grpc:%d"            // grpc:<userid>
	FollowingsKeyFormat = "grpc:%d:followings" // grpc:<userid>:followings
	FollowersKeyFormat  = "grpc:%d:followers"  // grpc:<userid>:follower
)

type (
	CacheDao struct {
		cfg CacheConfig

		redisCli *redis.Client
	}
	CacheConfig struct {
		Host string
		Port int
		TTL  time.Duration // not used now
	}
)

func New(cfg CacheConfig) (*CacheDao, error) {
	// TODO: validate config

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	redisCli := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := redisCli.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	dao := &CacheDao{
		cfg:      cfg,
		redisCli: redisCli,
	}
	return dao, nil
}

func (dao *CacheDao) SetCachedUser(ctx context.Context, user *model.User) error {
	if user == nil {
		return nil
	}
	data, err := json.Marshal(user)
	if err != nil {
		return err
	}

	key := getUserKey(user.ID)
	err = dao.redisCli.Set(ctx, key, string(data), 0).Err() // dont set TTL
	if err != nil {
		return err
	}
	return nil
}

func (dao *CacheDao) GetCachedUserByID(ctx context.Context, userId int64) (*model.User, error) {
	key := getUserKey(userId)

	data, err := dao.redisCli.Get(ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	user := &model.User{}
	err = json.Unmarshal([]byte(data), user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// AddCachedFollow add following to a sorted set: value=follower_id, score=timestamp
func (dao *CacheDao) AddCachedFollow(ctx context.Context, follow *model.Follow) error {
	logger.Debug("", logger.F("follow", follow))
	if follow == nil || follow.Follower == nil || follow.Following == nil {
		return errors.New("invalid follow data to cache")
	}

	// add to following key
	followingKey := getUserFollowingsKey(follow.Follower.ID)
	score := float64(follow.FollowTs)

	err := dao.redisCli.ZAdd(ctx, followingKey, redis.Z{
		Score:  score,
		Member: follow.Following.ID,
	}).Err()
	if err != nil {
		return err
	}

	// TODO: add to followers key also
	return nil
}

func (dao *CacheDao) GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error) {
	key := getUserFollowingsKey(userId)
	lastTs, ok := paging.LastValue.(int64)
	if !ok {
		return nil, errors.New("invalid last value")
	}

	// ZRevRangeByScore sort by score desc, means timestamp desc
	zs, err := dao.redisCli.ZRevRangeByScoreWithScores(ctx, key, &redis.ZRangeBy{
		Max:    strconv.Itoa(int(lastTs)),
		Offset: 0,
		Count:  paging.Limit,
	}).Result()
	if err != nil {
		return nil, err
	}

	// get ids and timestamps
	follows := make([]*CachedFollow, 0)
	for _, entry := range zs {
		member := entry.Member.(string)
		id, err := strconv.ParseInt(member, 10, 64)
		if err != nil {
			return nil, errors.New("failed to parse following id from cached")
		}
		ts := int64(entry.Score)
		follows = append(follows, &CachedFollow{
			ID:        id,
			Timestamp: ts,
		})
	}

	// get grpc data by ids
	userKeys := make([]string, len(follows))
	for i := range follows {
		userKeys[i] = fmt.Sprintf(UserKeyFormat, follows[i].ID)
	}
	datas, err := dao.redisCli.MGet(ctx, userKeys...).Result()
	if err != nil {
		return nil, err
	}

	followings := make([]*model.Follow, 0, len(follows))
	for i := range follows {
		userData, ok := datas[i].(string)
		if !ok {
			return nil, errors.New("failed to parse grpc data from cached")
		}
		user := &model.User{}
		err = json.Unmarshal([]byte(userData), user)
		if err != nil {
			return nil, err
		}

		followings = append(followings, &model.Follow{
			ID:        0, // we dont store in cache now ...
			Follower:  nil,
			Following: user,
			FollowTs:  follows[i].Timestamp,
		})
	}

	return followings, nil

}

func getUserKey(userId int64) string {
	return fmt.Sprintf(UserKeyFormat, userId)
}

func getUserFollowingsKey(userId int64) string {
	return fmt.Sprintf(FollowingsKeyFormat, userId)
}

func getUserFollowersKey(userId int64) string {
	return fmt.Sprintf(FollowersKeyFormat, userId)
}
