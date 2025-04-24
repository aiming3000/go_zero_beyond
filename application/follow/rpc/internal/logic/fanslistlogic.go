package logic

import (
	"context"
	"github.com/zeromicro/go-zero/core/threading"
	"go_zero_bryond/application/follow/rpc/code"
	"go_zero_bryond/application/follow/rpc/internal/model"
	"go_zero_bryond/application/follow/rpc/internal/types"
	"strconv"
	"time"

	"go_zero_bryond/application/follow/rpc/internal/svc"
	"go_zero_bryond/application/follow/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

const userfansExpireTime = 3600 * 24 * 2

type FansListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFansListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FansListLogic {
	return &FansListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 粉丝列表
func (l *FansListLogic) FansList1(in *pb.FansListRequest) (*pb.FansListResponse, error) {
	// todo: add your logic here and delete this line
	//第一步：参数校验处理
	if in.UserId <= 0 {
		return nil, code.UserIdEmpty
	}
	if in.PageSize <= 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		in.Cursor = time.Now().Unix()
	}
	//第二步：查询数据
	var (
		err            error
		isCache, isEnd bool
		lastId, cursor int64
		fansedUIds     []int64
		fans           []*model.Follow
		curPage        []*pb.FansItem
	)

	fansUIds, createTime, _ := l.cacheFansUserIds(l.ctx, in.UserId, in.Cursor, in.PageSize)
	if len(fansUIds) > 0 { //缓存中有数据
		isCache = true
		if fansUIds[len(fansUIds)-1] == -1 {
			fansUIds = fansUIds[:len(fansUIds)-1]
			isEnd = true
		}
		if len(fansUIds) == 0 {
			return &pb.FansListResponse{}, err
		}
		fansedUIds = fansUIds
		for i, fansUId := range fansUIds {
			curPage = append(curPage, &pb.FansItem{
				UserId:     in.UserId,
				FansUserId: fansUId,
				CreateTime: createTime[i],
			})
		}

	} else { // 缓存中没有数据查询数据库
		fans, err = l.svcCtx.FollowModel.FindByFollowedUserId(l.ctx, in.UserId, types.CacheMaxFansCount)
		if err != nil {
			l.Logger.Errorf("[FansList] FollowModel.FindByFollowedUserId error: %v req: %v", err, in)
			return nil, err
		}
		if len(fans) == 0 {
			return &pb.FansListResponse{}, nil
		}
		var firstPageFans []*model.Follow
		if len(fans) >= int(in.PageSize) {
			firstPageFans = fans[:in.PageSize]
		} else {
			firstPageFans = fans
			isEnd = true
		}
		for _, fan := range firstPageFans {
			fansedUIds = append(fansedUIds, fan.UserID)
			curPage = append(curPage, &pb.FansItem{
				UserId:     fan.UserID,
				FansUserId: fan.FollowedUserID,
				CreateTime: fan.CreateTime.Unix(),
			})
		}

	}
	if len(curPage) > 0 {
		pagelast := curPage[len(curPage)-1]
		lastId = pagelast.FansUserId
		cursor = pagelast.CreateTime
		if cursor < 0 {
			cursor = 0
		}
		for k, fan := range curPage {
			if fan.CreateTime == in.Cursor && fan.FansUserId == in.Id {
				curPage = curPage[k:]
				break
			}
		}
	}
	fa, err := l.svcCtx.FollowCountModel.FindByUserIds(l.ctx, fansedUIds)
	if err != nil {
		l.Logger.Errorf("[FansList] FollowCountModel.FindByUserIds error: %v fansedUIds: %v", err, fansedUIds)
	}
	uidFansCount := make(map[int64]int)
	uidFollowCount := make(map[int64]int)
	for _, f := range fa {
		uidFansCount[f.UserID] = f.FansCount
		uidFollowCount[f.UserID] = f.FollowCount
	}
	for _, cur := range curPage {
		cur.FansCount = int64(uidFansCount[cur.UserId])
		cur.FollowCount = int64(uidFollowCount[cur.UserId])
	}

	//uidCount := make(map[int64]*pb.FansItem)
	//for _, f := range fa {
	//	uidCount[f.UserID].FansCount = int64(f.FansCount)
	//	uidCount[f.UserID].FansCount = int64(f.FollowCount)
	//}
	//for _, cur := range curPage {
	//	cur.FansCount = uidCount[cur.UserId].FansCount
	//	cur.FollowCount = uidCount[cur.UserId].FollowCount
	//}
	//异步写入缓存
	if !isCache {
		threading.GoSafe(func() {
			if len(fans) < types.CacheMaxFollowCount && len(fans) > 0 {
				fans = append(fans, &model.Follow{UserID: -1})
			}
			err := l.addCacheFans(context.Background(), in.UserId, fans)
			if err != nil {
				logx.Errorf("addCacheFans error: %v", err)
			}
		})
	}
	//第三步：返回结果处理
	return &pb.FansListResponse{
		Items:  curPage,
		Cursor: cursor,
		IsEnd:  isEnd,
		Id:     lastId,
	}, nil
}

// FansList 粉丝列表.
func (l *FansListLogic) FansList(in *pb.FansListRequest) (*pb.FansListResponse, error) {
	if in.UserId == 0 {
		return nil, code.UserIdEmpty
	}
	if in.PageSize == 0 {
		in.PageSize = types.DefaultPageSize
	}
	if in.Cursor == 0 {
		in.Cursor = time.Now().Unix()
	}
	var (
		err            error
		isCache, isEnd bool
		lastId, cursor int64
		fansUserIds    []int64
		fansModel      []*model.Follow
		curPage        []*pb.FansItem
	)
	fansUIds, createTime, _ := l.cacheFansUserIds(l.ctx, in.UserId, in.Cursor, in.PageSize)
	if len(fansUIds) > 0 {
		isCache = true
		if fansUIds[len(fansUIds)-1] == -1 {
			fansUIds = fansUIds[:len(fansUIds)-1]
			isEnd = true
		}
		if len(fansUIds) == 0 {
			return &pb.FansListResponse{}, nil
		}
		fansUserIds = fansUIds
		for i, fansUId := range fansUIds {
			curPage = append(curPage, &pb.FansItem{
				UserId:     in.UserId,
				FansUserId: fansUId,
				CreateTime: createTime[i],
			})
		}
	} else {
		fansModel, err = l.svcCtx.FollowModel.FindByFollowedUserId(l.ctx, in.UserId, types.CacheMaxFansCount)
		if err != nil {
			l.Logger.Errorf("[FansList] FollowModel.FindByFollowedUserId error: %v req: %v", err, in)
			return nil, err
		}
		if len(fansModel) == 0 {
			return &pb.FansListResponse{}, nil
		}
		var firstPageFans []*model.Follow
		if len(fansModel) > int(in.PageSize) {
			firstPageFans = fansModel[:in.PageSize]
		} else {
			firstPageFans = fansModel
			isEnd = true
		}
		for _, fans := range firstPageFans {
			fansUserIds = append(fansUserIds, fans.UserID)
			curPage = append(curPage, &pb.FansItem{
				UserId:     fans.FollowedUserID,
				FansUserId: fans.UserID,
				CreateTime: fans.CreateTime.Unix(),
			})
		}
	}
	if len(curPage) > 0 {
		pageLast := curPage[len(curPage)-1]
		lastId = pageLast.FansUserId
		cursor = pageLast.CreateTime
		if cursor < 0 {
			cursor = 0
		}
		for i, fans := range curPage {
			if fans.CreateTime == in.Cursor && fans.FansUserId == in.Id {
				curPage = curPage[i:]
				break
			}
		}
	}
	fa, err := l.svcCtx.FollowCountModel.FindByUserIds(l.ctx, fansUserIds)
	if err != nil {
		l.Logger.Errorf("[FansList] FollowCountModel.FindByUserIds error: %v fansUserIds: %v", err, fansUserIds)
	}
	uidFansCount := make(map[int64]int)
	uidFollowCount := make(map[int64]int)
	for _, f := range fa {
		uidFansCount[f.UserID] = f.FansCount
		uidFollowCount[f.UserID] = f.FollowCount
	}
	for _, cur := range curPage {
		cur.FansCount = int64(uidFansCount[cur.FansUserId])
		cur.FollowCount = int64(uidFollowCount[cur.FansUserId])
	}

	ret := &pb.FansListResponse{
		Items:  curPage,
		Cursor: cursor,
		IsEnd:  isEnd,
		Id:     lastId,
	}

	if !isCache {
		threading.GoSafe(func() {
			if len(fansModel) < types.CacheMaxFansCount && len(fansModel) > 0 {
				fansModel = append(fansModel, &model.Follow{UserID: -1})
			}
			err = l.addCacheFans(context.Background(), in.UserId, fansModel)
		})
	}
	return ret, nil
}

func (l *FansListLogic) cacheFansUserIds(ctx context.Context, userId, cursor, pageSize int64) ([]int64, []int64, error) {
	key := userFansKey(userId)
	b, err := l.svcCtx.BizRedis.ExistsCtx(l.ctx, key)
	if err != nil {
		logx.Errorf("[cacheFansUserIds] BizRedis.ExistsCtx error: %v", err)
	}
	if b {
		err = l.svcCtx.BizRedis.ExpireCtx(ctx, key, userfansExpireTime)
		if err != nil {
			logx.Errorf("[cacheFansUserIds] BizRedis.ExpireCtx error: %v", err)
		}
	}

	pairs, err := l.svcCtx.BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx(ctx, key, 0, cursor, 0, int(pageSize))
	if err != nil {
		logx.Errorf("[cacheFansUserIds] BizRedis.ZrevrangebyscoreWithScoresAndLimitCtx error: %v", err)
		return nil, nil, err
	}
	var uids []int64
	var createTimes []int64
	for _, pair := range pairs {
		uid, err := strconv.ParseInt(pair.Key, 10, 64)
		if err != nil {
			logx.Errorf("[cacheFansUserIds] strconv.ParseInt userId error: %v", err)
			continue
		}
		uids = append(uids, uid)
		createTime, err := strconv.ParseInt(strconv.FormatInt(pair.Score, 10), 10, 64)
		if err != nil {
			logx.Errorf("[cacheFansUserIds] strconv.ParseInt createTime error: %v", err)
			continue
		}
		createTimes = append(createTimes, createTime)
	}

	return uids, createTimes, nil
}

func (l *FansListLogic) addCacheFans(ctx context.Context, userId int64, fans []*model.Follow) error {
	if len(fans) == 0 {
		return nil
	}
	key := userFansKey(userId)
	for _, fan := range fans {
		var score int64
		if fan.UserID == -1 {
			score = 0
		} else {
			score = fan.CreateTime.Unix()
		}
		_, err := l.svcCtx.BizRedis.ZaddCtx(ctx, key, score, strconv.FormatInt(fan.UserID, 10))
		if err != nil {
			return err
		}

	}

	return l.svcCtx.BizRedis.ExpireCtx(ctx, key, userfansExpireTime)
}
