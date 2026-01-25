package service

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
)

// LotteryDetailHandler 活动详情接口
func LotteryDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	ids := r.URL.Query().Get("id")
	id, err := strconv.Atoi(ids)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid ID"
		writeJSON(w, res)
		return
	}

	lottery, err := dao.Imp.GetLotteryByID(int32(id))
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = lottery
	writeJSON(w, res)
}

// DrawRequest 抽奖请求
type DrawRequest struct {
	LotteryID int32 `json:"lotteryId"`
}

// DrawResponse 抽奖响应
type DrawResponse struct {
	Success   bool   `json:"success"`
	PrizeName string `json:"prizeName"`
	IsWon     bool   `json:"isWon"`
}

// LotteryDrawHandler 抽奖接口
func LotteryDrawHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// 1. 鉴权
	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	// 2. 解析参数
	decoder := json.NewDecoder(r.Body)
	var req DrawRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// 3. 获取活动
	lottery, err := dao.Imp.GetLotteryByID(req.LotteryID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	if lottery.Status != 1 {
		res.Code = -1
		res.ErrorMsg = "Activity ended"
		writeJSON(w, res)
		return
	}

	// 4. 检查积分 (假设消耗10积分)
	cost := 10
	if user.Points < cost {
		res.Code = -1
		res.ErrorMsg = "Insufficient points"
		writeJSON(w, res)
		return
	}

	// 5. 抽奖算法
	rand.Seed(time.Now().UnixNano())
	isWon := rand.Float64() < lottery.Probability
	prizeName := "Thank you for participating"
	if isWon {
		prizeName = "iPhone 16" // 简化处理，实际应从配置获取
	}

	// 6. 更新数据 (事务处理最好，这里简化为顺序操作)
	// 扣减积分
	user.Points -= cost
	if err := dao.Imp.UpsertUser(user); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to update points"
		writeJSON(w, res)
		return
	}

	// 记录抽奖
	record := &model.UserLotteryRecord{
		UserID:    user.ID,
		LotteryID: lottery.ID,
		PrizeName: prizeName,
		CreatedAt: time.Now(),
	}
	if err := dao.Imp.CreateRecord(record); err != nil {
		// Log error (points already deducted...)
	}

	// 7. 返回结果
	res.Code = 0
	res.Data = DrawResponse{
		Success:   true,
		PrizeName: prizeName,
		IsWon:     isWon,
	}
	writeJSON(w, res)
}
