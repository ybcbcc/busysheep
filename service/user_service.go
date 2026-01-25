package service

import (
	"math"
	"net/http"
	"time"

	"wxcloudrun-golang/db/dao"
)

// UserInfoHandler 用户信息接口
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = user
	writeJSON(w, res)
}

// UserLotteryHistoryHandler 抽奖历史接口
func UserLotteryHistoryHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	records, err := dao.Imp.GetUserRecords(user.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch records"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = records
	writeJSON(w, res)
}

// UserPublishHistoryHandler 发布历史接口
func UserPublishHistoryHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	posts, err := dao.Imp.GetUserPosts(user.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch posts"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = posts
	writeJSON(w, res)
}

// MemberInfoHandler 会员信息接口
func MemberInfoHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	var days int
	if user.IsMember && user.MemberExpireAt.After(time.Now()) {
		days = int(math.Ceil(user.MemberExpireAt.Sub(time.Now()).Hours() / 24))
	} else {
		days = 0
	}

	res.Code = 0
	res.Data = map[string]interface{}{
		"isMember":      user.IsMember,
		"daysRemaining": days,
		"expireAt":      user.MemberExpireAt,
	}
	writeJSON(w, res)
}
