package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go_zero_bryond/application/user/rpc/internal/config"
	"go_zero_bryond/application/user/rpc/internal/model"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(conn, c.CacheRedis),
	}
}
