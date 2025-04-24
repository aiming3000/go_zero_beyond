package code

import "go_zero_bryond/pkg/xcode"

var (
	FollowUserIdEmpty   = xcode.New(40001, "关注用户id为空")
	FollowedUserIdEmpty = xcode.New(40002, "被关注用户id为空")
	CannotFollowSelf    = xcode.New(40003, "不能关注自己")
	UserIdEmpty         = xcode.New(40004, "用户id为空")
	IsFollowed          = xcode.New(40005, "已关注，不需要再次关注")
)
