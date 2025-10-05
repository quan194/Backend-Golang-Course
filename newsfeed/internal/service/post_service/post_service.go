package post_service

import (
	"context"
	"time"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type PostDAI interface {
}

type PostCacheDAI interface {
}

type UserCacheDAI interface {
}

type PostMsgProducer interface {
	SendPost(ctx context.Context, post *model.Post) error
}

type PostService struct {
	dai             PostDAI
	userCacheDai    UserCacheDAI
	postCacheDai    PostCacheDAI
	postMsgProducer PostMsgProducer
}

func New(postDai PostDAI, userCacheDai UserCacheDAI, postCacheDai PostCacheDAI, postMsgProducer PostMsgProducer) (*PostService, error) {
	svc := &PostService{
		dai:             postDai,
		userCacheDai:    userCacheDai,
		postCacheDai:    postCacheDai,
		postMsgProducer: postMsgProducer,
	}

	return svc, nil
}

func (s *PostService) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	// 1. validate grpc
	// 2. validate post
	// 3. create post into database
	// 4. send msg to kafka: s.postMsgProducer.SendPost(ctx, post)

	// test producer
	if err := s.postMsgProducer.SendPost(ctx, &model.Post{
		ID:               9999,
		UserID:           9999,
		Content:          "demo",
		CreatedTimestamp: time.Now().Unix(),
	}); err != nil {
		logger.Error("failed to send post", logger.E(err))
	} else {
		logger.Debug("send post ok")
	}
	return nil, nil
}

func (s *PostService) GetPostByUserID(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error) {
	// flow is same as get followings:
	// 1. get post_ids by page
	// 2. get post models by post_ids
	return nil, common.NewError(common.CodeNotImplemented, "Not Implemented")
}

func (s *PostService) GetNewsfeed(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error) {
	// flow is same as get followings
	// 1. get post_ids by page
	// 2. get post models by post_ids
	return nil, common.NewError(common.CodeNotImplemented, "Not Implemented")
}

func (s *PostService) AppendPostToNewsfeed(ctx context.Context, post *model.Post) error {
	// 1. create cached post: grpc:<user_id>:post:<post_id>
	// 2. get all follower_ids from cache key grpc:<post_user_id>:followers (should get by batch)
	// 3. add post_id + timestamp to sorted set grpc:<followerid>:newsfeed
	return common.NewError(common.CodeNotImplemented, "Not Implemented")
}
