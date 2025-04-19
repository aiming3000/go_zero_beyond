package logic

import (
	"context"
	"go_zero_bryond/application/article/rpc/internal/code"
	"go_zero_bryond/application/article/rpc/internal/types"

	"go_zero_bryond/application/article/rpc/internal/svc"
	"go_zero_bryond/application/article/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleDeleteLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArticleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDeleteLogic {
	return &ArticleDeleteLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArticleDeleteLogic) ArticleDelete(in *pb.ArticleDeleteRequest) (*pb.ArticleDeleteResponse, error) {
	// todo: add your logic here and delete this line
	//logx.Errorf("UpdateArticleStatus req: %v error: %v", in, errors.New("log 测试内容"))
	//l.Logger.Errorf("UpdateArticleStatus req: %v error: %v", in, errors.New("log 测试内容"))  //会多trace
	//return nil, code.UserIdInvalid
	//第一步：参数校验处理
	if in.UserId <= 0 {
		return nil, code.UserIdInvalid
	}
	if in.ArticleId <= 0 {
		return nil, code.ArticleIdInvalid
	}

	// 第二步调用model 处理数据
	err := l.svcCtx.ArticleModel.UpdateArticleStatus(l.ctx, in.ArticleId, types.ArticleStatusUserDelete)
	if err != nil {
		l.Logger.Errorf("UpdateArticleStatus req: %v error: %v", in, err)
		return nil, err
	}

	// 处理缓存数据 为了列表缓存Redis和mysql数据保持一致
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, articlesKey(in.UserId, types.SortPublishTime), in.ArticleId)
	if err != nil {
		l.Logger.Errorf("ZremCtx req: %v error: %v", in, err)
	}
	_, err = l.svcCtx.BizRedis.ZremCtx(l.ctx, articlesKey(in.UserId, types.SortLikeCount), in.ArticleId)
	if err != nil {
		l.Logger.Errorf("ZremCtx req: %v error: %v", in, err)
	}

	// 第三步返回数据处理
	return &pb.ArticleDeleteResponse{}, nil
}
