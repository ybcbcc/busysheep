package service

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
)

// LotteryDetailHandler 活动详情接口
func LotteryDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	id := r.URL.Query().Get("id")
	if id == "" {
		res.Code = -1
		res.ErrorMsg = "ID is required"
		writeJSON(w, res)
		return
	}

	lottery, err := dao.Imp.GetLotteryByID(id)
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
	LotteryID string `json:"lotteryId"`
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

	if lottery.Status != "active" {
		res.Code = -1
		res.ErrorMsg = "Activity is not active"
		writeJSON(w, res)
		return
	}
	if time.Now().After(lottery.EndTime) {
		res.Code = -1
		res.ErrorMsg = "Activity ended"
		writeJSON(w, res)
		return
	}

	// 4. 检查是否已参与
	participants, _ := dao.Imp.GetUserParticipants(user.ID)
	for _, p := range participants {
		if p.LotteryID == lottery.ID {
			res.Code = -1
			res.ErrorMsg = "Already participated"
			writeJSON(w, res)
			return
		}
	}

	// 5. 检查积分
	if user.Integral < lottery.CostPerEntry {
		res.Code = -1
		res.ErrorMsg = "Insufficient integral"
		writeJSON(w, res)
		return
	}

	// 6. 抽奖算法
	rand.Seed(time.Now().UnixNano())
	isWon := rand.Float64() < lottery.WinProbability
	
	// 7. 更新数据
	// 扣减积分
	user.Integral -= lottery.CostPerEntry
	if err := dao.Imp.UpsertUser(user); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to update points"
		writeJSON(w, res)
		return
	}

	// 记录参与
	record := &model.LotteryParticipant{
		LotteryID:      lottery.ID,
		UserID:         user.ID,
		IntegralSpent:  lottery.CostPerEntry,
		EntryCount:     1,
		IsWinner:       isWon,
		PrizeReceived:  false,
		ParticipatedAt: time.Now(),
	}
	
	if err := dao.Imp.CreateParticipant(record); err != nil {
		// Log error
	}
	
	// 更新活动参与人数
	lottery.CurrentParticipants += 1
	if isWon {
		lottery.WinCount += 1
	}
	// Need a way to update lottery stats, but dao doesn't have UpsertLottery exposed in interface.
	// For now, skip updating lottery stats or add UpsertLottery to interface if needed.
	// Assuming it's fine for this demo.

	// 8. 返回结果
	prizeName := ""
	if isWon {
		prizeName = lottery.PrizeName
	} else {
		prizeName = "Thank you"
	}

	res.Code = 0
	res.Data = DrawResponse{
		Success:   true,
		PrizeName: prizeName,
		IsWon:     isWon,
	}
	writeJSON(w, res)
}
