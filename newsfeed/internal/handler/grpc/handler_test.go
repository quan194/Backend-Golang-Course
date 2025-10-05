package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"

	user_pb "ep.k16/newsfeed/internal/handler/proto/grpc"
	"ep.k16/newsfeed/internal/service/model"
)

func TestUserGrpcHandler_Signup(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockService := new(MockUserService)
		handler := &userGrpcHandler{userService: mockService}

		req := &user_pb.SignupRequest{
			UserName:    proto.String("username"),
			Password:    proto.String("password"),
			DisplayName: proto.String("displayname"),
			Email:       proto.String("email@gmail.com"),
			Dob:         proto.String("19990101"),
		}

		expectedUser := &model.User{
			ID:          1,
			Username:    "username",
			DisplayName: "displayname",
			Email:       "email@gmail.com",
			Dob:         "19990101",
		}

		// Expect the service.Signup call
		mockService.On("Signup", mock.Anything, mock.MatchedBy(func(u *model.User) bool {
			return u.Username == "username" && u.Password == "password"
		})).Return(expectedUser, nil).Once()

		resp, err := handler.Signup(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedUser.ID, resp.User.GetId())
		assert.Equal(t, expectedUser.Username, resp.User.GetUserName())
		assert.Equal(t, expectedUser.DisplayName, resp.User.GetDisplayName())
		assert.Equal(t, expectedUser.Email, resp.User.GetEmail())
		assert.Equal(t, expectedUser.Dob, resp.User.GetDob())
	})
}
