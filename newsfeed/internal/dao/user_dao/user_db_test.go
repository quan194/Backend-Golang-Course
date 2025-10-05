package user_dao

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ep.k16/newsfeed/internal/service/model"
)

func TestUserDAI_Create(t *testing.T) {
	// Assume
	// mock gorm db with sqlmock
	sqlDB, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer sqlDB.Close()

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	assert.NoError(t, err)

	dai := &UserDAI{db: gormDB}

	user := &model.User{
		Username:       "username",
		HashedPassword: "hashed123",
		Email:          "email@gmail.com",
		DisplayName:    "User",
		Dob:            "20000101",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO").
		WithArgs(
			user.Username,
			user.HashedPassword,
			user.Email,
			user.DisplayName,
			user.Dob,
			false, // removed
		).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	// Act
	created, err := dai.Create(context.Background(), user)

	// Assẻt
	assert.NoError(t, err)
	user.ID = 1
	assert.Equal(t, user, created)

	err = mock.ExpectationsWereMet()
	assert.NoError(t, err)
}
