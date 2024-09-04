package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"go_zero_bryond/application/article/api/internal/config"
	"go_zero_bryond/application/article/rpc/article"
	"go_zero_bryond/application/user/rpc/user"
)

type ServiceContext struct {
	Config     config.Config
	ArticleRPC article.Article
	UserRPC    user.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:     c,
		ArticleRPC: article.NewArticle(zrpc.MustNewClient(c.ArticleRPC)),
		UserRPC:    user.NewUser(zrpc.MustNewClient(c.UserRPC)),
	}
}
