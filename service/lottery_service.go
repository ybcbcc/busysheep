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
	// 结束条件：参与截止或抽奖时间结束
	now := time.Now()
	afterDeadline := now.After(lottery.ParticipationDeadline)
	if afterDeadline || now.After(lottery.EndTime) {
		// 到期后进行结算，保证奖品抽完（若仍有剩余）
		finalizeRemainingPrizes(lottery)
		res.Code = -1
		res.ErrorMsg = "Activity ended"
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

	// 6. 抽奖算法：动态概率，保证奖品抽完
	rand.Seed(time.Now().UnixNano())
	remainingPrizes := lottery.PrizeQuantity - lottery.WinCount
	isWon := false
	if remainingPrizes > 0 {
		// 计算剩余名额（包含当前这一次）
		remainingSlots := lottery.MaxParticipants - lottery.CurrentParticipants
		if remainingSlots < 1 {
			remainingSlots = 1
		}
		// 顺序抽取m个中奖者算法：每次中奖概率 = 剩余奖品数 / 剩余参与名额
		p := float64(remainingPrizes) / float64(remainingSlots)
		if p < 0 {
			p = 0
		}
		if p > 1 {
			p = 1
		}
		isWon = rand.Float64() < p
	}
	
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
	// 更新展示用概率：进行中使用动态概率，闭合后固定为 奖品数量/总参与人数
	if lottery.CurrentParticipants >= lottery.MaxParticipants {
		if lottery.CurrentParticipants > 0 {
			lottery.WinProbability = float64(lottery.PrizeQuantity) / float64(lottery.CurrentParticipants)
		}
		lottery.Status = "finished"
		lottery.ActualDrawTime = &now
	} else {
		// 动态概率（剩余奖品 / 剩余名额）
		nextRemainingPrizes := lottery.PrizeQuantity - lottery.WinCount
		nextRemainingSlots := lottery.MaxParticipants - lottery.CurrentParticipants
		if nextRemainingSlots < 1 {
			nextRemainingSlots = 1
		}
		lottery.WinProbability = float64(nextRemainingPrizes) / float64(nextRemainingSlots)
	}
	
	if err := dao.Imp.UpdateLottery(lottery); err != nil {
		fmt.Printf("Failed to update lottery stats: %v\n", err)
		// Don't fail the request as the draw was successful
	}

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
	lottery.ActualDrawTime = &t
	_ = dao.Imp.UpdateLottery(lottery)
}
