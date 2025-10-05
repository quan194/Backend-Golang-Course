package user_service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ep.k16/newsfeed/internal/common"
	"ep.k16/newsfeed/internal/service/model"
)

func Test_checkPassword(t *testing.T) {
	rawPassword := "123456"

	hashedPassword, err := hashPassword(rawPassword)
	assert.NoError(t, err)

	matched := checkPassword(hashedPassword, rawPassword)
	assert.True(t, matched)
}

func assertAppError(t *testing.T, err error, errCode common.ErrorCode) {
	appError, ok := err.(*common.AppError)
	assert.True(t, ok)
	assert.NotNil(t, appError)
	assert.Equal(t, appError.Code, errCode)
}

func TestUserService_Signup(t *testing.T) {
	ctx := context.Background()

	t.Run("db error when checking username", func(t *testing.T) {
		user := &model.User{
			ID:             0,
			Username:       "username",
			Password:       "password",
			HashedPassword: "xxxxxxxx",
			DisplayName:    "display_name",
			Email:          "abc@gmail.com",
			Dob:            "19900101",
		}

		mockDAI := new(MockUserDAI)
		mockDAI.On("GetByUsername", ctx, "username").Return((*model.User)(nil), errors.New("db down"))

		service := &UserService{dai: mockDAI}

		res, err := service.Signup(ctx, user)

		assert.Nil(t, res)
		assert.Error(t, err)
		assertAppError(t, err, common.CodeDatabaseError)
	})

	t.Run("username already exists", func(t *testing.T) {
		user := &model.User{
			ID:             0,
			Username:       "username",
			Password:       "password",
			HashedPassword: "",
			DisplayName:    "display_name",
			Email:          "abc@gmail.com",
			Dob:            "19900101",
		}

		mockDAI := new(MockUserDAI)
		mockDAI.On("GetByUsername", ctx, "username").Return(&model.User{
			ID:             1,
			Username:       "username",
			Password:       "",
			HashedPassword: "xxxxxxxx",
			DisplayName:    "display_name",
			Email:          "abc@gmail.com",
			Dob:            "19900101",
		}, nil)

		service := &UserService{dai: mockDAI}
		res, err := service.Signup(ctx, user)

		assert.Nil(t, res)
		assert.Error(t, err)
		assertAppError(t, err, common.CodeExistedUsername)
	})

	t.Run("success with cache enabled", func(t *testing.T) {
		user := &model.User{
			ID:             0,
			Username:       "username",
			Password:       "password",
			HashedPassword: "",
			DisplayName:    "display_name",
			Email:          "abc@gmail.com",
			Dob:            "19900101",
		}

		mockDAI := new(MockUserDAI)
		mockDAI.On("GetByUsername", ctx, "username").Return((*model.User)(nil), nil)
		mockDAI.On("Create", ctx, mock.AnythingOfType("*model.User")).
			Return(&model.User{
				ID:             1,
				Username:       "username",
				Password:       "",
				HashedPassword: "xxxxxxxx",
				DisplayName:    "display_name",
				Email:          "abc@gmail.com",
				Dob:            "19900101",
			}, nil)

		mockCache := new(MockUserCacheDAI)
		mockCache.On("SetCachedUser", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		service := &UserService{dai: mockDAI, cacheDai: mockCache, enabledCache: true}

		res, err := service.Signup(ctx, user)

		assert.NoError(t, err)
		assert.Equal(t, int64(1), res.ID)
	})
}
