package logic

import (
	"context"
	"fmt"
	"go_zero_bryond/application/article/api/internal/code"
	"go_zero_bryond/application/article/api/internal/svc"
	"go_zero_bryond/application/article/api/internal/types"
	"go_zero_bryond/application/article/rpc/article"
	"go_zero_bryond/application/user/rpc/user"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArticleDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArticleDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArticleDetailLogic {
	return &ArticleDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

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
