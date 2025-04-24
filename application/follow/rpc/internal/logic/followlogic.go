package logic

import (
	"context"
	"fmt"
	"go_zero_bryond/application/follow/rpc/code"
	"go_zero_bryond/application/follow/rpc/internal/model"
	"go_zero_bryond/application/follow/rpc/internal/types"
	"gorm.io/gorm"
	"strconv"
	"time"

	"go_zero_bryond/application/follow/rpc/internal/svc"
	"go_zero_bryond/application/follow/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 关注
func (l *FollowLogic) Follow(in *pb.FollowRequest) (*pb.FollowResponse, error) {
	// todo: add your logic here and delete this line
	if in.UserId == 0 {
		return nil, code.FollowUserIdEmpty
	}
	if in.FollowedUserId == 0 {
		return nil, code.FollowedUserIdEmpty
	}
	if in.UserId == in.FollowedUserId {
		return nil, code.CannotFollowSelf
	}

	// 关注逻辑：第一种情况 初次关注   第二种情况  取消后再次关注

	// 查询数据是否关注过
	is_follow, err := l.svcCtx.FollowModel.FindByUserIDAndFollowedUserID(l.ctx, in.UserId, in.FollowedUserId)
	if err != nil {
		l.Logger.Errorf("FindByUserIDAndFollowedUserID is error: %v", err)
		return nil, err
	}

	if is_follow != nil { // 处理取消后再次关注逻辑
		if is_follow.ID >= 0 && is_follow.FollowStatus == types.FollowStatusFollow { //关注状态不需要再关注
			return nil, code.IsFollowed
		}
		err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			err := model.NewFollowModel(tx).UpdateFields(l.ctx, is_follow.ID, map[string]interface{}{
				"follow_status": types.FollowStatusFollow,
				"UpdateTime":    time.Now(),
			})
			if err != nil {
				return err
			}
			err = model.NewFollowCountModel(tx).IncrFollowCount(l.ctx, in.UserId)
			if err != nil {
				return err
			}
			return model.NewFollowCountModel(tx).IncrFansCount(l.ctx, in.FollowedUserId)
		})
		if err != nil {
			l.Logger.Errorf("[Follow] Transaction error: %v", err)
			return nil, err
		}

	} else { // 处理首次关注逻辑
		// 事务
		err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			err := model.NewFollowModel(tx).Insert(l.ctx, &model.Follow{
				UserID:         in.UserId,
				FollowedUserID: in.FollowedUserId,
				FollowStatus:   types.FollowStatusFollow,
				CreateTime:     time.Now(),
				UpdateTime:     time.Now(),
			})
			if err != nil {
				return err
			}
			err = model.NewFollowCountModel(tx).IncrFollowCount(l.ctx, in.UserId)
			if err != nil {
				return err
			}
			return model.NewFollowCountModel(tx).IncrFansCount(l.ctx, in.FollowedUserId)
		})
		if err != nil {
			l.Logger.Errorf("[Follow] Transaction error: %v", err)
			return nil, err
		}
	}

	//缓存数据处理
	followExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFollowKey(in.UserId))
	if err != nil {
		l.Logger.Errorf("[Follow] Redis Exists error: %v", err)
		return nil, err
	}
	if followExist {
		_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, userFollowKey(in.UserId), time.Now().Unix(), strconv.FormatInt(in.FollowedUserId, 10))
		if err != nil {
			l.Logger.Errorf("[Follow] Redis Zadd error: %v", err)
			return nil, err
		}
		_, err = l.svcCtx.BizRedis.ZremrangebyrankCtx(l.ctx, userFollowKey(in.UserId), 0, -(types.CacheMaxFollowCount + 1))
		if err != nil {
			l.Logger.Errorf("[Follow] Redis Zremrangebyrank error: %v", err)
		}
	}
	fansExist, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, userFansKey(in.FollowedUserId))
	if err != nil {
		l.Logger.Errorf("[Follow] Redis Exists error: %v", err)
		return nil, err
	}
	if fansExist {
		_, err = l.svcCtx.BizRedis.ZaddCtx(l.ctx, userFansKey(in.FollowedUserId), time.Now().Unix(), strconv.FormatInt(in.UserId, 10))
		if err != nil {
			l.Logger.Errorf("[Follow] Redis Zadd error: %v", err)
			return nil, err
		}
		_, err = l.svcCtx.BizRedis.ZremrangebyrankCtx(l.ctx, userFansKey(in.FollowedUserId), 0, -(types.CacheMaxFansCount + 1))
		if err != nil {
			l.Logger.Errorf("[Follow] Redis Zremrangebyrank error: %v", err)
		}
	}

	return &pb.FollowResponse{}, nil
}

func userFollowKey(userId int64) string {
	return fmt.Sprintf("biz#user#follow#%d", userId)
}

func userFansKey(userId int64) string {
	return fmt.Sprintf("biz#user#fans#%d", userId)
}
