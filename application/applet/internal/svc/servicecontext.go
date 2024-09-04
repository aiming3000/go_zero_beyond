package svc

import (
	"github.com/zeromicro/go-zero/zrpc"
	"go_zero_bryond/application/applet/internal/config"
	"go_zero_bryond/application/user/rpc/user"
)

type ServiceContext struct {
	Config  config.Config
	UserRPC user.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 自定义拦截器
	//userRPC := zrpc.MustNewClient(c.UserRPC, zrpc.WithUnaryClientInterceptor(interceptors.ClientErrorInterceptor()))

	return &ServiceContext{
		Config: c,
		//UserRPC: user.NewUser(userRPC),
		UserRPC: user.NewUser(zrpc.MustNewClient(c.UserRPC)),
	}
}
