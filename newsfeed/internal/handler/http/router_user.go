package http

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/proto"

	"ep.k16/newsfeed/internal/common"
	grpc_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/pkg/logger"
)

type FollowRequest struct {
	PeerId int64 `json:"peer_id"`
}

type FollowData struct {
	Follower        *UserData `json:"follower,omitempty"`  // optional
	Following       *UserData `json:"following,omitempty"` // optional
	FollowTimestamp int64     `json:"follow_ts"`
	FollowTime      string    `json:"follow_time"`
}

func (h *Server) Follow(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req = &FollowRequest{}
		api = c.Request.Method + " " + c.Request.RequestURI

		userId = c.GetInt64("user_id")
	)

	// bind req
	if err := c.ShouldBind(req); err != nil {
		bindErr := common.WrapError(common.CodeInvalidRequest, "bind request error", err)
		h.returnErrResp(c, bindErr)
		return
	}
	logger.Debug("parse request", logger.F("api", api), logger.F("req", req))

	// validate req etc ...

	// process logic
	grpcReq := &grpc_pb.FollowRequest{
		UserId: proto.Int64(userId),
		PeerId: proto.Int64(req.PeerId),
	}

	grpcResp, err := h.grpcClient.Follow(ctx, grpcReq)
	if err != nil {
		appErr := common.FromGRPCError(err)
		h.returnErrResp(c, appErr)
		return
	}

	// process response
	data := &FollowData{}
	data.Following = &UserData{
		ID:          grpcResp.GetFollowing().GetId(),
		Username:    grpcResp.GetFollowing().GetUserName(),
		Email:       grpcResp.GetFollowing().GetEmail(),
		DisplayName: grpcResp.GetFollowing().GetDisplayName(),
		Dob:         grpcResp.GetFollowing().GetDob(),
	}

	h.returnDataResp(c, "Follow successfully", data)
}

type GetFollowersRequest struct {
	// paging
}

func (h *Server) GetFollowers(c *gin.Context) {

}

type GetFollowingsRequest struct {
	Limit     int64 `json:"limit"`
	LastValue int64 `json:"last_value"`
}

type FollowingsData struct {
	Followings []*FollowData `json:"followings"`
}

func (h *Server) GetFollowings(c *gin.Context) {
	var (
		ctx    = c.Request.Context()
		api    = c.Request.Method + " " + c.Request.RequestURI
		userId = c.GetInt64("user_id")
	)

	// bind query param
	req, err := parseGetFollowingsRequest(c)
	if err != nil {
		h.returnErrResp(c, err)
		return
	}

	logger.Debug("parse request", logger.F("api", api), logger.F("req", req))

	// validate req etc ...

	// process logic
	grpcReq := &grpc_pb.GetFollowingsRequest{
		UserId: proto.Int64(userId),
		Paging: &grpc_pb.FollowPaging{
			Limit:     proto.Int64(req.Limit),
			LastValue: proto.Int64(req.LastValue),
		},
	}

	grpcResp, err := h.grpcClient.GetFollowings(ctx, grpcReq)
	if err != nil {
		appErr := common.FromGRPCError(err)
		h.returnErrResp(c, appErr)
		return
	}

	// process response
	data := &FollowingsData{}
	for _, f := range grpcResp.GetFollowings() {
		data.Followings = append(data.Followings, &FollowData{
			Following: &UserData{
				ID:          f.GetFollowing().GetId(),
				Username:    f.GetFollowing().GetUserName(),
				Email:       f.GetFollowing().GetEmail(),
				DisplayName: f.GetFollowing().GetDisplayName(),
				Dob:         f.GetFollowing().GetDob(),
			},
			FollowTimestamp: f.GetFollowTimestamp(),
			FollowTime:      time.Unix(f.GetFollowTimestamp(), 0).Format("2006-01-02 15:04:05"),
		})
	}

	h.returnDataResp(c, "Get followings successfully", data)
}

func parseGetFollowingsRequest(c *gin.Context) (*GetFollowingsRequest, error) {
	limit, err := strconv.Atoi(c.Query("limit"))
	if err != nil {
		return nil, common.NewError(common.CodeInvalidRequest, "invalid limit query")
	}

	lastValue, err := strconv.Atoi(c.Query("last_value"))
	if err != nil {
		return nil, common.NewError(common.CodeInvalidRequest, "invalid last_value query")
	}

	return &GetFollowingsRequest{
		Limit:     int64(limit),
		LastValue: int64(lastValue),
	}, nil
}
