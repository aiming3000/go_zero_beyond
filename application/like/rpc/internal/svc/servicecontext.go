package svc

import (
	"github.com/zeromicro/go-queue/kq"
	"go_zero_bryond/application/like/rpc/internal/config"
)

type ServiceContext struct {
	Config         config.Config
	KqPusherClient *kq.Pusher
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		KqPusherClient: kq.NewPusher(c.KqPusherConf.Brokers, c.KqPusherConf.Topic),
	}
}
