package http

import (
	"github.com/gin-gonic/gin"

	"ep.k16/newsfeed/internal/common"
	grpc_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/pkg/logger"
)

type CreatePostRequest struct {
}

type PostData struct {
}

func (h *Server) CreatePost(c *gin.Context) {
	var (
		ctx = c.Request.Context()
		req = &CreatePostRequest{}
		api = c.Request.Method + " " + c.Request.RequestURI

		// userId = c.GetInt64("user_id")
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
	grpcReq := &grpc_pb.CreatePostRequest{}

	_, err := h.grpcClient.CreatePost(ctx, grpcReq)
	if err != nil {
		appErr := common.FromGRPCError(err)
		h.returnErrResp(c, appErr)
		return
	}

	// process response
	h.returnDataResp(c, "Create post successfully", map[string]interface{}{"TODO": "TODO"})
}
