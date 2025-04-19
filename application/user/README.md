# 用户服务

## 用户服务介绍
 用户RPC服务
*     用户注册
*     用户登录
*     获取用户信息

#User RPC 服务搭建
## 首先编写proto文件
```
syntax = "proto3";

package service;
option go_package="./service";

service User {
  rpc Register(RegisterRequest) returns (RegisterResponse);
  rpc FindById(FindByIdRequest) returns (FindByIdResponse);
  rpc FindByMobile(FindByMobileRequest) returns (FindByMobileResponse);
  rpc SendSms(SendSmsRequest) returns (SendSmsResponse);
}


message RegisterRequest {
  string username = 1;
  string mobile = 2;
  string avatar = 3;
  string password = 4;
}

message RegisterResponse {
  int64 userId = 1;
}

message FindByIdRequest {
  int64 userId = 1;
}

message FindByIdResponse {
  int64 userId = 1;
  string username = 2;
  string mobile = 3;
  string avatar = 4;
}

message FindByMobileRequest {
  string mobile = 1;
}

message FindByMobileResponse {
  int64 userId = 1;
  string username = 2;
  string mobile = 3;
  string avatar = 4;
}

message SendSmsRequest {
  int64 userId = 1;
  string mobile = 2;
}

message SendSmsResponse {
}
```
在.proto文件同级目录执行下面命令，执行proto文件
```
goctl rpc protoc ./user.proto --go_out=. --go-grpc_out=. --zrpc_out=./
```
## model模块书写
接下来 编写model内容 首先创建数据库，以及数据表
```sql
create database beyond_user;
use beyond_user;

CREATE TABLE `user` (
    `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `username` varchar(32) NOT NULL DEFAULT '' COMMENT '用户名',
    `avatar` varchar(256) NOT NULL DEFAULT '' COMMENT '头像',
    `mobile` varchar(128) NOT NULL DEFAULT '' COMMENT '手机号',
    `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_update_time` (`update_time`),
    UNIQUE KEY `uk_mobile` (`mobile`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='用户表';

insert into user(username, avatar, mobile) values ('张三', 'https://beyond-blog.oss-cn-beijing.aliyuncs.com/avatar/2021/01/01/1609488000.jpg', '13800138000');
```
在go_zero_beyond/application/user/rpc 目录下执行执行下面命令
```shell
goctl model mysql datasource --dir ./internal/model --table user --cache true --url "root:root@tcp(127.0.0.1:3306)/beyond_user"
```
## 数据库等配置修改

rpc/etc/user.yaml
```yaml
Name: user.rpc
ListenOn: 0.0.0.0:8090
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: user.rpc
  
#  新加数据库和Redis配置
DataSource: root:root@tcp(127.0.0.1:3306)/beyond_user?parseTime=true
CacheRedis:
  - Host: 127.0.0.1:6379
    Pass:
    Type: node
```

修改rpc/internal/config/config.go
```go
package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string
	CacheRedis cache.CacheConf
}

```


修改svc/servicecontext.go文件
```go
//原文件内容
package svc

import "rpc-user/internal/config"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}


// 修改为一下内容
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
```
至此配置等问题得以解决，书写逻辑

## 获取用户详情逻辑

修改go_zero_beyond/application/user/rpc/internal/logic/findbyidlogic.go

修改获取用户详情的逻辑
```go
func (l *FindByIdLogic) FindById(in *service.FindByIdRequest) (*service.FindByIdResponse, error) {
	// todo: add your logic here and delete this line
	user, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(in.UserId))
	if err != nil {
		logx.Errorf("FindById userId: %d error: %v", in.UserId, err)
		return nil, err
	}

	return &service.FindByIdResponse{
		UserId:   int64(user.Id),
		Username: user.Username,
		Avatar:   user.Avatar,
	}, nil
}
```
获取用户详情的逻辑结束，postman访问rpc示例
![](D:\ruanjian\Golang\go1.22.4\path\src\demo\go_zero_demo\go_zero_beyond\doc\image\img_1.png)
