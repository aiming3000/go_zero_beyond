package logic

import (
	"context"
	"errors"
	"go_zero_bryond/application/applet/internal/code"
	"go_zero_bryond/application/user/rpc/user"
	"go_zero_bryond/pkg/encrypt"
	"go_zero_bryond/pkg/jwt"
	"strings"

	"go_zero_bryond/application/applet/internal/svc"
	"go_zero_bryond/application/applet/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterRequest) (resp *types.RegisterResponse, err error) {
	// todo: add your logic here and delete this line
	req.Name = strings.TrimSpace(req.Name)
	req.Mobile = strings.TrimSpace(req.Mobile)
	if len(req.Mobile) == 0 {
		//return nil, errors.New("注册手机号不能为空")
		return nil, code.RegisterMobileEmpty
	}

	req.Password = strings.TrimSpace(req.Password)

	if len(req.Password) == 0 {
		return nil, errors.New("密码不能为空")
	} else {
		req.Password = encrypt.EncPassword(req.Password)
	}

	req.VerificationCode = strings.TrimSpace(req.VerificationCode)

	if len(req.VerificationCode) == 0 {
		return nil, errors.New("验证码不能为空")
	}

	//检测验证码逻辑
	//err := checkVerificationCode(l.svcCtx.BizRedis, req.Mobile, req.VerificationCode)
	//if err != nil {
	//	logx.Errorf("checkVerificationCode error: %v", err)
	//	return nil, err
	//}

	mobile, err := encrypt.EncMobile(req.Mobile)
	if err != nil {
		logx.Errorf("EncMobile mobile: %s error: %v", req.Mobile, err)
		return nil, err
	}
	//调用user RPC服务
	u, err := l.svcCtx.UserRPC.FindByMobile(l.ctx, &user.FindByMobileRequest{
		Mobile: mobile,
	})
	if err != nil {
		logx.Errorf("FindByMobile error: %v", err)
		return nil, err
	}
	if u != nil && u.UserId > 0 {
		logx.Errorf("FindByMobile error: %v", errors.New("该手机号已经注册"))
		return nil, errors.New("该手机号已经注册")
	}

	regRet, err := l.svcCtx.UserRPC.Register(l.ctx, &user.RegisterRequest{
		Username: req.Name,
		Mobile:   mobile,
	})
	if err != nil {
		logx.Errorf("Register error: %v", err)
		return nil, err
	}
	token, err := jwt.BuildTokens(jwt.TokenOptions{
		AccessSecret: l.svcCtx.Config.Auth.AccessSecret,
		AccessExpire: l.svcCtx.Config.Auth.AccessExpire,
		Fields: map[string]interface{}{
			"userId": regRet.UserId,
		},
	})
	if err != nil {
		logx.Errorf("BuildTokens error: %v", err)
		return nil, err
	}
	//_ = delActivationCache(req.Mobile, req.VerificationCode, l.svcCtx.BizRedis)

	return &types.RegisterResponse{
		UserId: regRet.UserId,
		Token: types.Token{
			AccessToken:  token.AccessToken,
			AccessExpire: token.AccessExpire,
		},
	}, nil
}
