package service

import (
	"encoding/json"
	"fmt"
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

	// 时间诊断日志（北京时间）
	bj := time.FixedZone("CST", 8*3600)
	serverNow := time.Now().In(bj)
	fmt.Printf("[TimeDiag][detail][BJ] now=%s start=%s end=%s status=%s\n",
		serverNow.Format("2006-01-02 15:04:05"), lottery.StartTime.In(bj).Format("2006-01-02 15:04:05"), lottery.EndTime.In(bj).Format("2006-01-02 15:04:05"), lottery.Status)

	res.Code = 0
	res.Data = lottery
	writeJSON(w, res)
}

// DrawRequest 抽奖请求
type DrawRequest struct {
	LotteryID string `json:"lotteryId"`
	ClientNow string `json:"clientNow"`
}

// DrawResponse 抽奖响应
type DrawResponse struct {
	Success   bool   `json:"success"`
	PrizeName string `json:"prizeName"`
	IsWon     bool   `json:"isWon"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ResultResponse struct {
	Success   bool   `json:"success"`
	IsWon     bool   `json:"isWon"`
	PrizeName string `json:"prizeName"`
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
	// 使用客户端时间进行比较（北京时间日志）
	bj := time.FixedZone("CST", 8*3600)
	clientNow := time.Now().In(bj)
	if req.ClientNow != "" {
		if t, err := time.Parse(time.RFC3339, req.ClientNow); err == nil {
			clientNow = t.In(bj)
		}
	}
	fmt.Printf("[TimeDiag][draw][BJ] now=%s start=%s end=%s before_start=%t after_end=%t\n",
		clientNow.Format("2006-01-02 15:04:05"),
		lottery.StartTime.In(bj).Format("2006-01-02 15:04:05"), lottery.EndTime.In(bj).Format("2006-01-02 15:04:05"),
		clientNow.Before(lottery.StartTime), clientNow.After(lottery.EndTime))
	if clientNow.Before(lottery.StartTime) {
		res.Code = -1
		res.ErrorMsg = "Not started"
		writeJSON(w, res)
		return
	}
	if clientNow.After(lottery.EndTime) {
		// 到期后进行统一开奖，并允许已报名用户查看结果
		finalizeRemainingPrizes(lottery)
		// 查找当前用户报名记录
		participants, _ := dao.Imp.GetUserParticipants(user.ID)
		var mine *model.LotteryParticipant
		for _, p := range participants {
			if p.LotteryID == lottery.ID {
				mine = p
				break
			}
		}
		if mine == nil {
			res.Code = -1
			res.ErrorMsg = "Not participated"
			writeJSON(w, res)
			return
		}
		res.Code = 0
		res.Data = ResultResponse{
			Success:   true,
			IsWon:     mine.IsWinner,
			PrizeName: func() string { if mine.IsWinner { return lottery.PrizeName } else { return "" } }(),
		}
		writeJSON(w, res)
		return
	}
	
	// Check max participants
	if lottery.CurrentParticipants >= lottery.MaxParticipants {
		// 达到人数上限后也进行结算以保证奖品抽完（最后一名在本次请求中处理）
		res.Code = -1
		res.ErrorMsg = "Participants limit reached"
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

	// 6. 报名：扣积分并记录参与，不立即开奖
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
		IsWinner:       false,
		PrizeReceived:  false,
		ParticipatedAt: clientNow,
	}
	
	if err := dao.Imp.CreateParticipant(record); err != nil {
		// Log error
	}
	
	// 更新活动参与人数
	lottery.CurrentParticipants += 1
	// 报名期不更新中奖统计，截止后统一开奖
	
	if err := dao.Imp.UpdateLottery(lottery); err != nil {
		fmt.Printf("Failed to update lottery stats: %v\n", err)
		// Don't fail the request as the draw was successful
	}

	// 7. 返回报名成功
	res.Code = 0
	res.Data = RegisterResponse{Success: true, Message: "Registered"}
	writeJSON(w, res)
}

// finalizeRemainingPrizes 到期结算：将剩余奖品随机分配给未中奖参与者，保证奖品抽完
func finalizeRemainingPrizes(lottery *model.Lottery) {
	remaining := lottery.PrizeQuantity - lottery.WinCount
	if remaining <= 0 {
		return
	}
	participants, err := dao.Imp.GetLotteryParticipants(lottery.ID)
	if err != nil || len(participants) == 0 {
		return
	}
	// 收集未中奖用户
	var losers []*model.LotteryParticipant
	for _, p := range participants {
		if !p.IsWinner {
			losers = append(losers, p)
		}
	}
	if len(losers) == 0 {
		return
	}
	// 随机打乱
	rand.Shuffle(len(losers), func(i, j int) { losers[i], losers[j] = losers[j], losers[i] })
	// 选取前remaining个置为中奖
	count := remaining
	if count > len(losers) {
		count = len(losers)
	}
	for i := 0; i < count; i++ {
		losers[i].IsWinner = true
		// 标记发奖待领取
	}
	// 批量保存
	for i := 0; i < count; i++ {
		_ = dao.Imp.UpdateParticipant(losers[i])
	}
	lottery.WinCount += count
	// 固定最终概率
	if lottery.CurrentParticipants > 0 {
		lottery.WinProbability = float64(lottery.PrizeQuantity) / float64(lottery.CurrentParticipants)
	}
	lottery.Status = "finished"
	t := time.Now()
	needReward := lottery.ActualDrawTime == nil
	lottery.ActualDrawTime = &t
	_ = dao.Imp.UpdateLottery(lottery)
	if needReward {
		owner, err := dao.Imp.GetUserByID(lottery.CreatorID)
		if err == nil && owner != nil {
			owner.Integral += 10
			_ = dao.Imp.UpsertUser(owner)
		}
	}
}
