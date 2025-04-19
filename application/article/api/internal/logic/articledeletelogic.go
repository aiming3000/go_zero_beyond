package logic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go_zero_bryond/application/article/api/internal/code"
	"go_zero_bryond/application/article/rpc/article"
	"go_zero_bryond/pkg/xcode"
	"strconv"

	"go_zero_bryond/application/article/api/internal/svc"
	"go_zero_bryond/application/article/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleDeleteLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleDeleteLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDeleteLogic {
	return &ArticleDeleteLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArticleDeleteLogic) ArticleDelete(req *types.ArticleDetailRequest) (resp *types.ArticleDetailResponse, err error) {
	// 第一步：参数处理
	if int(req.ArticleId) <= 0 {
		return nil, code.ArticleIdEmpty
	}
	// 获取当前用户ID
	userId, err := l.ctx.Value("userId").(json.Number).Int64()
	if err != nil {
		logx.Errorf("l.ctx.Value error: %v", err)
		return nil, xcode.NoLogin
	}
	// 第二步：调用RPC获取数据
	//调用RPC 获取文章详情  检查是否是当前用户数据
	articleInfo, err := l.svcCtx.ArticleRPC.ArticleDetail(l.ctx, &article.ArticleDetailRequest{ArticleId: req.ArticleId})
	if err != nil {
		logx.Errorf("get article detail id: %d err: %v", req.ArticleId, err)
		return nil, err
	}
	if articleInfo == nil || articleInfo.Article == nil {
		return nil, nil
	}
	if articleInfo.Article.AuthorId != userId {
		return nil, errors.New("不是改用户发布，无法删除")
	}

	// 若为当前用户发布文章则调用RPC删除接口删除
	deleteRes, err := l.svcCtx.ArticleRPC.ArticleDelete(l.ctx, &article.ArticleDeleteRequest{
		ArticleId: req.ArticleId,
		UserId:    userId,
	})
	if err != nil {
		logx.Errorf("delete article is fail article id: %d ; err: %v", req.ArticleId, err)
		return nil, err
	}
	fmt.Println(deleteRes)

	// 第三步：处理返回数据
	return &types.ArticleDetailResponse{
		Title:    articleInfo.Article.Title,
		Content:  articleInfo.Article.Content,
		Cover:    articleInfo.Article.Cover,
		AuthorId: strconv.FormatInt(articleInfo.Article.AuthorId, 10),
	}, nil
}
