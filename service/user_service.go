package service

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sort"
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

// UpdateUserRequest 更新用户信息请求
type UpdateUserRequest struct {
	Nickname  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

// UpdateUserHandler 更新用户信息接口
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// 1. Auth
	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	// 2. Parse
	decoder := json.NewDecoder(r.Body)
	var req UpdateUserRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// 3. Update
	if req.Nickname != "" {
		user.Nickname = req.Nickname
	}
	if req.AvatarURL != "" {
		user.AvatarURL = req.AvatarURL
	}
	user.UpdatedAt = time.Now()

	if err := dao.Imp.UpsertUser(user); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to update user: %v", err)
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

	// 1. 获取旧的 Post 发布
	posts, err := dao.Imp.GetUserPosts(user.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch posts"
		writeJSON(w, res)
		return
	}

	// 2. 获取新的 Lottery 发布
	lotteries, err := dao.Imp.GetUserLotteries(user.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch lotteries"
		writeJSON(w, res)
		return
	}

	// 3. 混合并排序
	// 复用 HomeItem 结构，因为前端列表页逻辑类似
	type HistoryItem struct {
		Type      string      `json:"type"` // "lottery" or "post"
		Data      interface{} `json:"data"`
		CreatedAt time.Time   `json:"createdAt"`
	}

	items := make([]HistoryItem, 0)
	for _, l := range lotteries {
		items = append(items, HistoryItem{Type: "lottery", Data: l, CreatedAt: l.CreatedAt})
	}
	for _, p := range posts {
		items = append(items, HistoryItem{Type: "post", Data: p, CreatedAt: p.CreatedAt})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	res.Code = 0
	res.Data = items
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
	if user.IsMember && user.MemberExpireAt != nil && user.MemberExpireAt.After(time.Now()) {
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
