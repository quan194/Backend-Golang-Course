package post_dao

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"ep.k16/newsfeed/internal/service/model"
)

type (
	PostDAO struct {
		db *gorm.DB
	}

	PostDbConfig struct {
		Username     string
		Password     string
		Host         string
		Port         int
		DatabaseName string
	}
)

func New(conf *PostDbConfig) (*PostDAO, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		conf.Username, conf.Password, conf.Host, conf.Port, conf.DatabaseName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %s", err)
	}

	return &PostDAO{
		db: db,
	}, nil
}

func (d *PostDAO) CreatePost(ctx context.Context, post *model.Post) (*model.Post, error) {
	panic("implement me")
}
