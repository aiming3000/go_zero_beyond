package logic

import (
	"context"
	"encoding/json"
	"errors"
	"go_zero_bryond/application/article/api/internal/code"
	"go_zero_bryond/application/article/rpc/pb"
	"go_zero_bryond/pkg/xcode"
	"strings"

	"go_zero_bryond/application/article/api/internal/svc"
	"go_zero_bryond/application/article/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

const minContentLen = 80

type PublishLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPublishLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PublishLogic {
	return &PublishLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

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
