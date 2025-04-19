package logic

import (
	"context"
	"go_zero_bryond/application/article/rpc/internal/code"
	"go_zero_bryond/application/article/rpc/internal/model"
	"go_zero_bryond/application/article/rpc/internal/types"
	"strconv"
	"time"

	"go_zero_bryond/application/article/rpc/internal/svc"
	"go_zero_bryond/application/article/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type PublishLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *PublishLogic) Publish(in *pb.PublishRequest) (*pb.PublishResponse, error) {
	// todo: add your logic here and delete this line
	// 第一步：参数校验处理
	if in.UserId <= 0 {
		return nil, code.UserIdInvalid
	}
	if len(in.Title) == 0 {
		return nil, code.ArticleTitleCantEmpty
	}
	if len(in.Content) == 0 {
		return nil, code.ArticleContentCantEmpty
	}

	// 第二部：利用model 操作数据库
	ret, err := l.svcCtx.ArticleModel.Insert(l.ctx, &model.Article{
		Title:       in.Title,
		Content:     in.Content,
		Cover:       in.Cover,
		Description: in.Description,
		AuthorId:    uint64(in.UserId),
		Status:      types.ArticleStatusVisible,
		PublishTime: time.Now(),
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	})
	if err != nil {
		l.Logger.Errorf("Publish Insert req: %v error: %v", in, err)
		return nil, err
	}
	articleId, err := ret.LastInsertId()
	if err != nil {
		l.Logger.Errorf("LastInsertId error: %v", err)
		return nil, err
	}

	// 为保证mysql和Redis数据一致性，加入更新缓存逻辑
	var (
		articleIdStr   = strconv.FormatInt(articleId, 10)
		publishTimeKey = articlesKey(in.UserId, types.SortPublishTime)
		likeNumkey     = articlesKey(in.UserId, types.SortLikeCount)
	)
	b, _ := l.svcCtx.BizRedis.ExistsCtx(l.ctx, publishTimeKey)
	if b {
		_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, publishTimeKey, time.Now().Unix(), articleIdStr)
		if err != nil {
			logx.Errorf("ZaddCtx req: %v error: %v", in, err)
		}
	}
	b, _ = l.svcCtx.BizRedis.ExistsCtx(l.ctx, likeNumkey)
	if b {
		_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, likeNumkey, 0, articleIdStr)
		if err != nil {
			logx.Errorf("ZaddCtx req: %v error: %v", in, err)
		}
	}

	// 第三步： 返回数据处理
	return &pb.PublishResponse{ArticleId: articleId}, nil

}
