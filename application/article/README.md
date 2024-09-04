# 文章服务
[https://pwmzlkcu3p.feishu.cn/docx/U9FGdVAFuoFFiUxySsgcl5TMnke](https://pwmzlkcu3p.feishu.cn/docx/U9FGdVAFuoFFiUxySsgcl5TMnke)

## 文章服务介绍
文章服务分为：


* api 为外部提供http接口服务
* admin 为后台文章审核功能
* mq 评论点赞功能交由mq处理
* rpc 提供rpc服务


## api服务搭建

在文章article目录下创建api目录
编写article.api 文件 go_zero_beyond/application/article/api/article.api

```
syntax = "v1"

type (
	UploadCoverResponse {
		CoverUrl string `json:"cover_url"`
	}
	PublishRequest {
		Title       string `json:"title"`
		Content     string `json:"content"`
		Description string `json:"description"`
		Cover       string `json:"cover"`
	}
	PublishResponse {
		ArticleId int64 `json:"article_id"`
	}
	ArticleDetailRequest {
		ArticleId int64 `form:"article_id"`
	}
	ArticleDetailResponse {
		Title       string `json:"title"`
		Content     string `json:"content"`
		Description string `json:"description"`
		Cover       string `json:"cover"`
		AuthorId    string `json:"author_id"`
		AuthorName  string `json:"author_name"`
	}
	ArticleListRequest {
		AuthorId  int64 `form:"author_id"`
		Cursor    int64 `form:"cursor"`
		PageSize  int64 `form:"page_size"`
		SortType  int32 `form:"sort_type"`
		ArticleId int64 `form:"article_id"`
	}
	ArticleInfo {
		ArticleId   int64  `json:"article_id"`
		Title       string `json:"title"`
		Content     string `json:"content"`
		Description string `json:"description"`
		Cover       string `json:"cover"`
	}
	ArticleListResponse {
		Articles []ArticleInfo `json:"articles"`
	}
)

@server (
	prefix: /v1/article
	//    signature: true
	jwt: Auth
)
service article-api {
	@handler UploadCoverHandler
	post /upload/cover returns (UploadCoverResponse)

	@handler PublishHandler
	post /publish (PublishRequest) returns (PublishResponse)

	@handler ArticleDetailHandler
	get /detail (ArticleDetailRequest) returns (ArticleDetailResponse)

	@handler ArticleListHandler
	get /list (ArticleListRequest) returns (ArticleListResponse)
}

//当前文件目录下
//goctl api go --dir=./  --api article.api

```

在api目录执行命令
```
goctl api go --dir=./  --api article.api
```

## rpc 服务搭建

首先编写proto文件 
go_zero_beyond/application/article/rpc/article.proto
```protobuf

syntax = "proto3";

package pb;
option go_package="./pb";

service Article {
  rpc Publish(PublishRequest) returns (PublishResponse);
  rpc Articles(ArticlesRequest) returns (ArticlesResponse);
  rpc ArticleDelete(ArticleDeleteRequest) returns (ArticleDeleteResponse);
  rpc ArticleDetail(ArticleDetailRequest) returns (ArticleDetailResponse);
}

message PublishRequest {
  int64 userId = 1;
  string title = 2;
  string content = 3;
  string description = 4;
  string cover = 5;
}

message PublishResponse {
  int64 articleId = 1;
}

message ArticlesRequest {
  int64 userId = 1;
  int64 cursor = 2;
  int64 pageSize = 3;
  int32 sortType = 4;
  int64 articleId = 5;
}

message ArticleItem {
  int64 Id = 1;
  string title = 2;
  string content = 3;
  string description = 4;
  string cover = 5;
  int64 commentCount = 6;
  int64 likeCount = 7;
  int64 publishTime = 8;
  int64 authorId = 9;
}

message ArticlesResponse {
  repeated ArticleItem articles = 1;
  bool isEnd = 2;
  int64 cursor = 3;
  int64 articleId = 4;
}

message ArticleDeleteRequest {
  int64 userId = 1;
  int64 articleId = 2;
}

message ArticleDeleteResponse {
}

message ArticleDetailRequest {
  int64 articleId = 1;
}

message ArticleDetailResponse {
  ArticleItem article = 1;
}

//.proto文件同级目录下执行
//goctl rpc protoc ./article.proto --go_out=. --go-grpc_out=. --zrpc_out=./
```

执行命名
```
goctl rpc protoc ./article.proto --go_out=. --go-grpc_out=. --zrpc_out=./
```

编写model内容
首先创建数据库，以及数据表
```sql
create database beyond_article;
use beyond_article;

CREATE TABLE `article` (
   `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
   `title` varchar(255) NOT NULL DEFAULT '' COMMENT '标题',
   `content` text COLLATE utf8_unicode_ci NOT NULL COMMENT '内容',
   `cover` varchar(255) NOT NULL DEFAULT '' COMMENT '封面',
   `description` varchar(255) NOT NULL DEFAULT '' COMMENT '描述',
   `author_id` bigint(20) UNSIGNED NOT NULL DEFAULT '0' COMMENT '作者ID',
   `status` tinyint(4) NOT NULL DEFAULT '0' COMMENT '状态 0:待审核 1:审核不通过 2:可见 3:用户删除',
   `comment_num` int(11) NOT NULL DEFAULT '0' COMMENT '评论数',
   `like_num` int(11) NOT NULL DEFAULT '0' COMMENT '点赞数',
   `collect_num` int(11) NOT NULL DEFAULT '0' COMMENT '收藏数',
   `view_num` int(11) NOT NULL DEFAULT '0' COMMENT '浏览数',
   `share_num` int(11) NOT NULL DEFAULT '0' COMMENT '分享数',
   `tag_ids` varchar(255) NOT NULL DEFAULT '' COMMENT '标签ID',
   `publish_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
   `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
   `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
   PRIMARY KEY (`id`),
   KEY `ix_author_id` (`author_id`),
   KEY `ix_update_time` (`update_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='文章表';


insert into article(title, content, author_id, like_num, publish_time) values ('文章标题1', '文章内容1', 1, 1, '2023-11-25 17:01:01');
insert into article(title, content, author_id, like_num, publish_time) values ('文章标题2', '文章内容2', 1, 10, '2023-11-25 15:01:01');

--数据表article已经建好，在go_zero_beyond/application/article/rpc 目录下执行
-- goctl model mysql datasource --dir ./internal/model --table article --cache true --url "root:root@tcp(127.0.0.1:3306)/beyond_article"
```
在go_zero_beyond/application/article/rpc 目录下执行执行下面命令
```
goctl model mysql datasource --dir ./internal/model --table article --cache true --url "root:root@tcp(127.0.0.1:3306)/beyond_article"
```

## api服务调用RPC服务

go_zero_beyond/application/applet/etc/applet-api.yaml

go_zero_beyond/application/applet/internal/config/config.go

go_zero_beyond/application/applet/internal/svc/servicecontext.go

在logic目录下的文件中调用，以获取文章详情为例

```go

func (l *ArticleDetailLogic) ArticleDetail(req *types.ArticleDetailRequest) (resp *types.ArticleDetailResponse, err error) {
	// todo: add your logic here and delete this line
	if req.ArticleId == 0 {
		return nil, errors.New("文章ID不能是0")
	}
	//调用ArticleRPC服务查询
	articleInfo, err := l.svcCtx.ArticleRPC.ArticleDetail(l.ctx, &article.ArticleDetailRequest{
		ArticleId: req.ArticleId,
	})
	
	if err != nil {
		logx.Errorf("get article detail id: %d err: %v", req.ArticleId, err)
		return nil, err
	}
	if articleInfo == nil || articleInfo.Article == nil {
		return nil, nil
	}
	//调用UserRPC服务查询
	userInfo, err := l.svcCtx.UserRPC.FindById(l.ctx, &user.FindByIdRequest{
		UserId: articleInfo.Article.AuthorId,
	})
	
	if err != nil {
		logx.Errorf("get userInfo id: %d err: %v", articleInfo.Article.AuthorId, err)
		return nil, err
	}
	return &types.ArticleDetailResponse{
		Title:       articleInfo.Article.Title,
		Content:     articleInfo.Article.Content,
		Description: articleInfo.Article.Description,
		Cover:       articleInfo.Article.Cover,
		AuthorId:    strconv.FormatInt(articleInfo.Article.AuthorId, 10),
		AuthorName:  userInfo.Username,
	}, nil
}
```

RPC服务查询数据库操作

go_zero_beyond/application/article/rpc/etc/article.yaml
```yaml
Name: article.rpc
ListenOn: 127.0.0.1:8081
Etcd:
  Hosts:
  - 127.0.0.1:2379
  Key: article.rpc
DataSource: root:root@tcp(127.0.0.1:3306)/beyond_article?parseTime=true
CacheRedis:
  - Host: 127.0.0.1:6379
    Pass:
    Type: node

```

go_zero_beyond/application/article/rpc/internal/config/config.go
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

go_zero_beyond/application/article/rpc/internal/svc/servicecontext.go
```go
package svc

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"go_zero_bryond/application/article/rpc/internal/config"
	"go_zero_bryond/application/article/rpc/internal/model"
)

type ServiceContext struct {
	Config config.Config
	ArticleModel model.ArticleModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	conn := sqlx.NewMysql(c.DataSource)
	return &ServiceContext{
		Config: c,
		ArticleModel: model.NewArticleModel(conn,c.CacheRedis),
	}
}

```
数据库配置连接成功，接下来具体调用。以获取文章详情为例
go_zero_beyond/application/article/rpc/internal/logic/articledetaillogic.go

```go
func (l *ArticleDetailLogic) ArticleDetail(in *pb.ArticleDetailRequest) (*pb.ArticleDetailResponse, error) {
	// todo: add your logic here and delete this line
	article, err := l.svcCtx.ArticleModel.FindOne(l.ctx, uint64(in.ArticleId))
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return &pb.ArticleDetailResponse{}, nil
		}
		return nil, err
	}
	return &pb.ArticleDetailResponse{
		Article: &pb.ArticleItem{
			Id:          int64(article.Id),
			Title:       article.Title,
			Content:     article.Content,
			Description: article.Description,
			Cover:       article.Cover,
			AuthorId:    int64(article.AuthorId),
			LikeCount:   article.LikeNum,
			PublishTime: article.PublishTime.Unix(),
		},
	}, nil
	
}
```
至此整个链路逻辑走通
重启所有服务userRPC、articleRPC、appletAPI、articleAPI利用postman调用测试得到如下结果
![](D:\ruanjian\Golang\go1.22.4\path\src\demo\go_zero_demo\go_zero_beyond\doc\image\img.png)

接下来完成其他RPC服务逻辑，这里不再赘述

