package service

import (
	"encoding/json"
	"fmt"
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

	// 获取参与记录
	records, err := dao.Imp.GetUserParticipants(user.ID)
	if err != nil {
		fmt.Printf("Error fetching participants for user %s: %v\n", user.ID, err)
		res.Code = -1
		res.ErrorMsg = "Failed to fetch records"
		writeJSON(w, res)
		return
	}
	fmt.Printf("Found %d participants for user %s\n", len(records), user.ID)

	// 为了前端展示方便，可能需要关联查询Lottery信息
	// 这里简单起见，返回 records，前端根据 lotteryId 可能需要二次查询或后端做聚合
	// 更好的做法是后端聚合，但 DAO 层目前分离。
	// 这里我们先返回 records，如果前端需要详情，可以再请求或这里做 Loop
	// 改进：返回一个包含 Lottery Title 的结构
	type HistoryItem struct {
		ID             int64     `json:"id"`
		LotteryID      string    `json:"lotteryId"`
		LotteryTitle   string    `json:"lotteryTitle"`
		IsWinner       bool      `json:"isWinner"`
		PrizeName      string    `json:"prizeName"`
		ParticipatedAt time.Time `json:"participatedAt"`
	}

	items := make([]HistoryItem, 0)
	for _, rec := range records {
		lottery, _ := dao.Imp.GetLotteryByID(rec.LotteryID)
		title := "Unknown Activity"
		if lottery != nil {
			title = lottery.Title
		}
		
		// Determine prize name (if winner, query lottery or store in record?)
		// The UserLotteryRecord in previous schema had PrizeName. The NEW LotteryParticipant DOES NOT have PrizeName.
		// It only has IsWinner. The PrizeName is in Lottery struct (Single prize per lottery in this simple schema).
		prizeName := ""
		if rec.IsWinner && lottery != nil {
			prizeName = lottery.PrizeName
		}

		items = append(items, HistoryItem{
			ID:             rec.ID,
			LotteryID:      rec.LotteryID,
			LotteryTitle:   title,
			IsWinner:       rec.IsWinner,
			PrizeName:      prizeName,
			ParticipatedAt: rec.ParticipatedAt,
		})
	}

	res.Code = 0
	res.Data = items
	writeJSON(w, res)
}

// UserPublishHistoryHandler 发布历史接口 (现在是发布的抽奖活动)
func UserPublishHistoryHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	lotteries, err := dao.Imp.GetUserCreatedLotteries(user.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch lotteries"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = lotteries
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
	isMember := user.MemberType != "free"
	if isMember && user.MemberExpiry != nil && user.MemberExpiry.After(time.Now()) {
		days = int(math.Ceil(user.MemberExpiry.Sub(time.Now()).Hours() / 24))
	} else {
		days = 0
	}

	res.Code = 0
	res.Data = map[string]interface{}{
		"memberType":    user.MemberType,
		"isMember":      isMember,
		"daysRemaining": days,
		"expireAt":      user.MemberExpiry,
	}
	writeJSON(w, res)
}
