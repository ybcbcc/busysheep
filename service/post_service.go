package service

import (
	"encoding/json"
	"net/http"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	"github.com/google/uuid"
)

// CreateLotteryRequest 发布抽奖请求 (原 CreatePostRequest)
type CreateLotteryRequest struct {
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	PrizeName       string  `json:"prizeName"`
	PrizeValue      int     `json:"prizeValue"`
	PrizeType       string  `json:"prizeType"` // integral, etc.
	CostPerEntry    int     `json:"costPerEntry"`
	MaxParticipants int     `json:"maxParticipants"`
	WinProbability  float64 `json:"winProbability"`
	EndTime         string  `json:"endTime"` // "2023-01-01 12:00:00"
}

// CreatePostHandler 发布接口 (实际是创建 Lottery)
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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
	var req CreateLotteryRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// Parse Time
	endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
	if err != nil {
		// Try fallback format or default
		endTime = time.Now().Add(24 * time.Hour)
	}

	// 3. Save Lottery
	lottery := &model.Lottery{
		ID:              uuid.New().String(),
		CreatorID:       user.ID,
		Title:           req.Title,
		Description:     req.Description,
		PrizeName:       req.PrizeName,
		PrizeValue:      req.PrizeValue,
		PrizeType:       req.PrizeType,
		CostPerEntry:    req.CostPerEntry,
		MaxParticipants: req.MaxParticipants,
		WinProbability:  req.WinProbability,
		StartTime:       time.Now(),
		EndTime:         endTime,
		Status:          "active", // Default active for demo
		AuditStatus:     "approved", // Bypass audit
		IsPublic:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := dao.Imp.CreateLottery(lottery); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to create lottery"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = lottery
	writeJSON(w, res)
}
