package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	"github.com/google/uuid"
)

// CreateLotteryRequest 发布抽奖请求 (原 CreatePostRequest)
type CreateLotteryRequest struct {
	ID              string  `json:"id"` // 用于更新
	Title           string  `json:"title"`
	ImageURL        string  `json:"imageUrl"`
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
		ID:              strings.ReplaceAll(uuid.New().String(), "-", ""), // Remove hyphens
		CreatorID:       user.ID,
		Title:           req.Title,
		ImageURL:        req.ImageURL,
		Description:     req.Description,
		PrizeName:       req.PrizeName,
		PrizeValue:      req.PrizeValue,
		PrizeType:       req.PrizeType,
		CostPerEntry:    req.CostPerEntry,
		MaxParticipants: req.MaxParticipants,
		WinProbability:  req.WinProbability,
		StartTime:       time.Now(),
		EndTime:         endTime,
		Status:          "pending",
		AuditStatus:     "pending",
		IsPublic:        true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := dao.Imp.CreateLottery(lottery); err != nil {
		fmt.Printf("Error creating lottery: %v\n", err)
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to create lottery: %v", err)
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = lottery
	writeJSON(w, res)
}

// UpdatePostHandler 更新接口
func UpdatePostHandler(w http.ResponseWriter, r *http.Request) {
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

	if req.ID == "" {
		res.Code = -1
		res.ErrorMsg = "ID is required"
		writeJSON(w, res)
		return
	}

	// 3. Check Ownership
	lottery, err := dao.Imp.GetLotteryByID(req.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	if lottery.CreatorID != user.ID {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}

	// 4. Update fields
	if req.Title != "" { lottery.Title = req.Title }
	if req.Description != "" { lottery.Description = req.Description }
	if req.ImageURL != "" { lottery.ImageURL = req.ImageURL }
	if req.PrizeName != "" { lottery.PrizeName = req.PrizeName }
	if req.PrizeValue > 0 { lottery.PrizeValue = req.PrizeValue }
	if req.CostPerEntry > 0 { lottery.CostPerEntry = req.CostPerEntry }
	if req.MaxParticipants > 0 { lottery.MaxParticipants = req.MaxParticipants }
	if req.WinProbability > 0 { lottery.WinProbability = req.WinProbability }
	
	if req.EndTime != "" {
		endTime, err := time.Parse("2006-01-02 15:04:05", req.EndTime)
		if err == nil {
			lottery.EndTime = endTime
		}
	}
	
	lottery.UpdatedAt = time.Now()

	if err := dao.Imp.UpdateLottery(lottery); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to update lottery: %v", err)
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = lottery
	writeJSON(w, res)
}

// DeletePostHandler 删除接口
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// 1. Auth
	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	// 2. Parse ID
	id := r.URL.Query().Get("id")
	if id == "" {
		// Try JSON body
		decoder := json.NewDecoder(r.Body)
		var req struct { ID string `json:"id"` }
		if err := decoder.Decode(&req); err == nil {
			id = req.ID
		}
	}

	if id == "" {
		res.Code = -1
		res.ErrorMsg = "ID is required"
		writeJSON(w, res)
		return
	}

	// 3. Check Ownership
	lottery, err := dao.Imp.GetLotteryByID(id)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	if lottery.CreatorID != user.ID {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}

	// 4. Delete
	if err := dao.Imp.DeleteLottery(id); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to delete lottery: %v", err)
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.ErrorMsg = "Success"
	writeJSON(w, res)
}
