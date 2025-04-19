# Go实战干货五-项目结构说明 && 文件上传 && 对接阿里OSS && 文章发布 && 多服务联动测试
https://pwmzlkcu3p.feishu.cn/docx/EBzWdSFR5oPVMJxP1oOcUGojnTd

## 文件上传
### 阿里OSS注册
**OSS配置**
修改go_zero_demo/go_zero_beyond/application/article/api/etc/article-api.yaml
```yaml
Oss:
  Endpoint: oss-cn-shanghai.aliyuncs.com
  AccessKeyId: xxxxxxxxxxxxxxxxxxxx
  AccessKeySecret: xxxxxxxxxxxxxxxxxxxx
  BucketName: beyond-article
```
修改go_zero_demo/go_zero_beyond/application/article/api/internal/config/config.go

```go
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Oss        struct {
		Endpoint         string
		AccessKeyId      string
		AccessKeySecret  string
		BucketName       string
		ConnectTimeout   int64 `json:",optional"`
		ReadWriteTimeout int64 `json:",optional"`
	}
	ArticleRPC zrpc.RpcClientConf
	UserRPC    zrpc.RpcClientConf
	
}
```

修改go_zero_demo/go_zero_beyond/application/article/api/internal/svc/servicecontext.go
```go
package svc

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/zeromicro/go-zero/zrpc"
	"go_zero_bryond/application/article/api/internal/config"
	"go_zero_bryond/application/article/rpc/article"
	"go_zero_bryond/application/user/rpc/user"
)

const (
	defaultOssConnectTimeout   = 1
	defaultOssReadWriteTimeout = 3
)

type ServiceContext struct {
	Config     config.Config
	OssClient  *oss.Client
	ArticleRPC article.Article
	UserRPC    user.User
}

func NewServiceContext(c config.Config) *ServiceContext {
	if c.Oss.ConnectTimeout == 0 {
		c.Oss.ConnectTimeout = defaultOssConnectTimeout
	}
	if c.Oss.ReadWriteTimeout == 0 {
		c.Oss.ReadWriteTimeout = defaultOssReadWriteTimeout
	}
	oc, err := oss.New(c.Oss.Endpoint, c.Oss.AccessKeyId, c.Oss.AccessKeySecret,
		oss.Timeout(c.Oss.ConnectTimeout, c.Oss.ReadWriteTimeout))
	if err != nil {
		panic(err)
	}
	return &ServiceContext{
		Config:     c,
		OssClient:  oc,
		ArticleRPC: article.NewArticle(zrpc.MustNewClient(c.ArticleRPC)),
		UserRPC:    user.NewUser(zrpc.MustNewClient(c.UserRPC)),
	}
}

```
至此OSS相关配置完成

### 上传文件逻辑编写
修改go_zero_demo/go_zero_beyond/application/article/api/internal/logic/uploadcoverlogic.go

```go
package logic

import (
	"context"
	"fmt"
	"go_zero_bryond/application/article/api/internal/code"
	"net/http"
	"time"

	"go_zero_bryond/application/article/api/internal/svc"
	"go_zero_bryond/application/article/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const maxFileSize = 10 << 20 // 10MB

type UploadCoverLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadCoverLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadCoverLogic {
	return &UploadCoverLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadCoverLogic) UploadCover(req *http.Request) (resp *types.UploadCoverResponse, err error) {
	// todo: add your logic here and delete this line
	_ = req.ParseMultipartForm(maxFileSize)
	file, handler, err := req.FormFile("cover")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bucket, err := l.svcCtx.OssClient.Bucket(l.svcCtx.Config.Oss.BucketName)
	if err != nil {
		logx.Errorf("get bucket failed, err: %v", err)
		return nil, code.GetBucketErr
	}
	objectKey := genFilename(handler.Filename)
	err = bucket.PutObject(objectKey, file)
	if err != nil {
		logx.Errorf("put object failed, err: %v", err)
		return nil, code.PutBucketErr
	}

	return &types.UploadCoverResponse{CoverUrl: genFileURL(objectKey)}, nil
}
func genFilename(filename string) string {
	return fmt.Sprintf("%d_%s", time.Now().UnixMilli(), filename)
}

func genFileURL(objectKey string) string {
	return fmt.Sprintf("https://beyond-article-liu.oss-cn-beijing.aliyuncs.com/%s", objectKey)
}

```


修改go_zero_demo/go_zero_beyond/application/article/api/internal/handler/uploadcoverhandler.go
```go
func UploadCoverHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logic.NewUploadCoverLogic(r.Context(), svcCtx)
		resp, err := l.UploadCover(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}

```
![](D:\ruanjian\Golang\go1.22.4\path\src\demo\go_zero_demo\go_zero_beyond\doc\image\img_2.png)

## 文章功能

配置文件修改
go_zero_demo/go_zero_beyond/application/article/api/etc/article-api.yaml
```yaml
ArticleRpc:
  Etcd:
    Hosts:
      - 127.0.0.1:2379
    Key: article.rpc
```
go_zero_demo/go_zero_beyond/application/article/api/internal/config/config.go
```go
type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Oss struct {
		Endpoint         string
		AccessKeyId      string
		AccessKeySecret  string
		BucketName       string
		ConnectTimeout   int64 `json:",optional"`
		ReadWriteTimeout int64 `json:",optional"`
	}
	ArticleRPC zrpc.RpcClientConf  //配置文章RPC
}
```
修改go_zero_demo/go_zero_beyond/application/article/api/internal/svc/servicecontext.go
初始化article rpc client
```go
type ServiceContext struct {
   Config     config.Config
   OssClient  *oss.Client
   ArticleRPC article.Article
}

func NewServiceContext(c config.Config) *ServiceContext {
   if c.Oss.ConnectTimeout == 0 {
      c.Oss.ConnectTimeout = defaultOssConnectTimeout
   }
   if c.Oss.ReadWriteTimeout == 0 {
      c.Oss.ReadWriteTimeout = defaultOssReadWriteTimeout
   }
   oc, err := oss.New(c.Oss.Endpoint, c.Oss.AccessKeyId, c.Oss.AccessKeySecret,
      oss.Timeout(c.Oss.ConnectTimeout, c.Oss.ReadWriteTimeout))
   if err != nil {
      panic(err)
   }

   return &ServiceContext{
      Config:     c,
      OssClient:  oc,
      ArticleRPC: article.NewArticle(zrpc.MustNewClient(c.ArticleRPC)),
   }
}
```
### 文章发布
发布文章核心逻辑
go_zero_demo/go_zero_beyond/application/article/api/internal/logic/publishlogic.go
```go
func (l *PublishLogic) Publish(req *types.PublishRequest) (resp *types.PublishResponse, err error) {
	//第一步：数据处理验证
	req.Title = strings.TrimSpace(req.Title)
	if len(req.Title) == 0 {
		logx.Errorf("ArtitleTitleEmpty error: %v", errors.New("文章内容为空")) //加入日志
		return nil, code.ArtitleTitleEmpty
	}
	req.Content = strings.TrimSpace(req.Content)
	if len(req.Content) < minContentLen {
		return nil, code.ArticleContentTooFewWords
	}
	if len(req.Cover) == 0 {
		return nil, code.ArticleCoverEmpty
	}
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		logx.Errorf("l.ctx.Value error: %v", err)
		return nil, xcode.NoLogin
	}
	
	// 第二步：调用RPC接口
	pret, err := l.svcCtx.ArticleRPC.Publish(l.ctx, &pb.PublishRequest{
		UserId:      userId,
		Title:       req.Title,
		Content:     req.Content,
		Description: req.Description,
		Cover:       req.Cover,
	})
	if err != nil {
		logx.Errorf("l.svcCtx.ArticleRPC.Publish req: %v userId: %d error: %v", req, userId, err)
		return nil, err
	}
	
	// 第三步：返回数据
	return &types.PublishResponse{ArticleId: pret.ArticleId}, nil
}
```
articleRPC 逻辑这边不在赘述

### 文章详情
go_zero_demo/go_zero_beyond/application/article/api/internal/logic/articledetaillogic.go
```go
func (l *ArticleDetailLogic) ArticleDetail(req *types.ArticleDetailRequest) (resp *types.ArticleDetailResponse, err error) {
	//第一步： 参数校验处理
	if req.ArticleId == 0 {
		//return nil, errors.New("文章ID不能是0")
		return nil, code.ArticleIdEmpty
	}

	//第二步：调用ArticleRPC服务查询
	articleInfo, err := l.svcCtx.ArticleRPC.ArticleDetail(l.ctx, &article.ArticleDetailRequest{
		ArticleId: req.ArticleId,
	})
	fmt.Println(111)
	if err != nil {
		logx.Errorf("get article detail id: %d err: %v", req.ArticleId, err)
		return nil, err
	}
	fmt.Println(articleInfo)
	if articleInfo == nil || articleInfo.Article == nil {
		return nil, nil
	}
	userInfo, err := l.svcCtx.UserRPC.FindById(l.ctx, &user.FindByIdRequest{
		UserId: articleInfo.Article.AuthorId,
	})
	if err != nil {
		logx.Errorf("get userInfo id: %d err: %v", articleInfo.Article.AuthorId, err)
		return nil, err
	}
	// 第三步：返回数据处理
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

### 文章列表

