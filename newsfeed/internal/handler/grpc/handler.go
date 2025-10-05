package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"

	grpc_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/internal/service/model"
	"ep.k16/newsfeed/pkg/logger"
)

type UserService interface {
	Signup(ctx context.Context, user *model.User) (*model.User, error)
	Login(ctx context.Context, user *model.User) (*model.User, error)

	Follow(ctx context.Context, userId, peerId int64) (*model.Follow, error)
	Unfollow(ctx context.Context, userId, peerId int64) error
	GetFollowings(ctx context.Context, userId int64, paging *model.Paging) ([]*model.Follow, error)
}

type PostService interface {
	CreatePost(ctx context.Context, post *model.Post) (*model.Post, error)
	GetPostByUserID(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error)
	GetNewsfeed(ctx context.Context, userId int, paging model.Paging) ([]*model.Post, error)
}

type Config struct {
	Host string
	Port int
}

type GrpcServer struct {
	cfg Config

	grpcServer *grpc.Server
}

func New(cfg Config, userService UserService, postService PostService) (*GrpcServer, error) {
	s := &GrpcServer{
		cfg: cfg,
	}

	// init handler
	userHandler := &userGrpcHandler{
		userService: userService,
		postService: postService,
	}

	// register handler into grpc server
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(CustomizedInterceptor()),
	)
	grpc_pb.RegisterServiceServer(grpcServer, userHandler)

	s.grpcServer = grpcServer

	return s, nil
}

func (s *GrpcServer) Start() error {
	// open port
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen on addr", logger.E(err), logger.F("addr", addr))
		return fmt.Errorf("failed to listen on %s", addr)
	}

	// server listen from port
	logger.Info("grpc grpc server starting to serve ...", logger.F("addr", addr))
	err = s.grpcServer.Serve(lis)
	if err != nil {
		logger.Error("failed to serve on addr", logger.E(err), logger.F("addr", addr))
		return fmt.Errorf("failed to serve on %s", addr)
	}

	return nil
}

func (s *GrpcServer) Stop() {
	s.grpcServer.GracefulStop()
}

type userGrpcHandler struct {
	grpc_pb.UnimplementedServiceServer

	userService UserService
	postService PostService
}

func (h *userGrpcHandler) Signup(ctx context.Context, req *grpc_pb.SignupRequest) (*grpc_pb.SignupResponse, error) {
	createdUser, err := h.userService.Signup(ctx, &model.User{
		ID:          0,
		Username:    req.GetUserName(),
		Password:    req.GetPassword(),
		DisplayName: req.GetDisplayName(),
		Email:       req.GetEmail(),
		Dob:         req.GetDob(),
	})
	if err != nil {
		return nil, err
	}

	resp := &grpc_pb.SignupResponse{
		User: toUserPb(createdUser),
	}
	return resp, nil
}

func (h *userGrpcHandler) Login(ctx context.Context, req *grpc_pb.LoginRequest) (*grpc_pb.LoginResponse, error) {
	loginUser, err := h.userService.Login(ctx, &model.User{
		Username: req.GetUserName(),
		Password: req.GetPassword(),
	})
	if err != nil {
		return nil, err
	}

	resp := &grpc_pb.LoginResponse{
		User: toUserPb(loginUser),
	}
	return resp, nil
}

func (h *userGrpcHandler) Follow(ctx context.Context, req *grpc_pb.FollowRequest) (*grpc_pb.FollowResponse, error) {
	followData, err := h.userService.Follow(ctx, req.GetUserId(), req.GetPeerId())
	if err != nil {
		return nil, err
	}
	resp := &grpc_pb.FollowResponse{
		IsFollowed: proto.Bool(true),
		Pair: &grpc_pb.UserUserData{
			Id:          proto.Int64(followData.ID),
			FollowerId:  proto.Int64(followData.Follower.ID),
			FollowingId: proto.Int64(followData.Following.ID),
			FollowTs:    proto.Int64(followData.FollowTs),
		},
		Following: toUserPb(followData.Following),
	}
	return resp, nil
}

func (h *userGrpcHandler) Unfollow(ctx context.Context, req *grpc_pb.UnfollowRequest) (*grpc_pb.UnfollowResponse, error) {
	err := h.userService.Unfollow(ctx, req.GetUserId(), req.GetPeerId())
	if err != nil {
		return nil, err
	}
	resp := &grpc_pb.UnfollowResponse{
		IsUnfollowed: proto.Bool(true),
	}
	return resp, nil
}

func (h *userGrpcHandler) GetFollowers(context.Context, *grpc_pb.GetFollowersRequest) (*grpc_pb.GetFollowersResponse, error) {
	// TODO: implement
	return nil, nil
}

func (h *userGrpcHandler) GetFollowings(ctx context.Context, req *grpc_pb.GetFollowingsRequest) (*grpc_pb.GetFollowingsResponse, error) {
	followings, err := h.userService.GetFollowings(ctx, req.GetUserId(), &model.Paging{
		LastValue: req.GetPaging().GetLastValue(),
		Limit:     req.GetPaging().GetLimit(),
	})
	if err != nil {
		return nil, err
	}

	resp := &grpc_pb.GetFollowingsResponse{}
	for _, f := range followings {
		resp.Followings = append(resp.Followings, toFollowPb(f))
	}
	return resp, nil
}

func toUserPb(user *model.User) *grpc_pb.UserData {
	if user == nil {
		return nil
	}
	return &grpc_pb.UserData{
		Id:          proto.Int64(user.ID),
		UserName:    proto.String(user.Username),
		DisplayName: proto.String(user.DisplayName),
		Email:       proto.String(user.Email),
		Dob:         proto.String(user.Dob),
	}
}

func toFollowPb(follow *model.Follow) *grpc_pb.FollowData {
	return &grpc_pb.FollowData{
		Follower:        toUserPb(follow.Follower),
		Following:       toUserPb(follow.Following),
		FollowTimestamp: proto.Int64(follow.FollowTs),
	}
}

func (h *userGrpcHandler) CreatePost(ctx context.Context, req *grpc_pb.CreatePostRequest) (*grpc_pb.CreatePostResponse, error) {
	// TODO implement me
	h.postService.CreatePost(ctx, nil) // test

	return nil, nil
}

func (h *userGrpcHandler) GetPosts(ctx context.Context, req *grpc_pb.GetPostsRequest) (*grpc_pb.GetPostsResponse, error) {
	// TODO implement me
	panic("implement me")
}

func (h *userGrpcHandler) GetNewsfeed(ctx context.Context, req *grpc_pb.GetNewsfeedRequest) (*grpc_pb.GetNewsfeedResponse, error) {
	// TODO implement me
	panic("implement me")
}
