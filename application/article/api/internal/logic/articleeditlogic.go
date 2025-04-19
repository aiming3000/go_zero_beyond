package logic

import (
	"context"

	"go_zero_bryond/application/article/api/internal/svc"
	"go_zero_bryond/application/article/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleEditLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleEditLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleEditLogic {
	return &ArticleEditLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleEditLogic) ArticleEdit(req *types.PublishRequest) (resp *types.PublishResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
