package service

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var errShareRewardExists = errors.New("share reward already exists")

const (
	lotteryShareStatusPrepared = "prepared"
	lotteryShareStatusShared   = "shared"
)

type LotteryShareCreateRequest struct {
	LotteryID    string `json:"lotteryId"`
	ShareSource  string `json:"shareSource"`
	ShareChannel string `json:"shareChannel"`
	ShareCode    string `json:"shareCode"`
}

type LotteryShareCreateResponse struct {
	ShareCode   string `json:"shareCode"`
	ShareSource string `json:"shareSource"`
	ShareStatus string `json:"shareStatus"`
	Message     string `json:"message"`
}

type LotteryShareConfirmRequest struct {
	LotteryID   string `json:"lotteryId"`
	ShareCode   string `json:"shareCode"`
	ShareSource string `json:"shareSource"`
}

type LotteryShareConfirmResponse struct {
	ShareCode   string `json:"shareCode"`
	ShareSource string `json:"shareSource"`
	ShareStatus string `json:"shareStatus"`
	Message     string `json:"message"`
}

type LotteryShareRewardRequest struct {
	LotteryID   string `json:"lotteryId"`
	ShareUserID string `json:"shareUserId"`
	ShareCode   string `json:"shareCode"`
}

type LotteryShareRewardResponse struct {
	Rewarded        bool   `json:"rewarded"`
	Points          int    `json:"points"`
	Message         string `json:"message"`
	CurrentIntegral int    `json:"currentIntegral,omitempty"`
}

// LotteryShareCreateHandler 预生成分享票据
func LotteryShareCreateHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	var req LotteryShareCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	if req.LotteryID == "" {
		res.Code = -1
		res.ErrorMsg = "Lottery ID is required"
		writeJSON(w, res)
		return
	}

	if _, err := dao.Imp.GetLotteryByID(req.LotteryID); err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	shareCode := strings.TrimSpace(req.ShareCode)
	if shareCode == "" {
		shareCode = generateShareCode()
	}
	shareSource := strings.TrimSpace(req.ShareSource)
	if shareSource == "" {
		shareSource = "top_button"
	}
	shareChannel := strings.TrimSpace(req.ShareChannel)
	if shareChannel == "" {
		shareChannel = "wechat_friend"
	}

	record := &model.LotteryShareRecord{
		LotteryID:    req.LotteryID,
		ShareUserID:  user.ID,
		ShareCode:    shareCode,
		ShareSource:  shareSource,
		ShareChannel: shareChannel,
		ShareStatus:  lotteryShareStatusPrepared,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := dao.Imp.CreateLotteryShareRecord(record); err != nil {
		// 分享成功回调可能因网络重试重复提交，遇到同一个分享码时直接复用即可。
		existing, getErr := dao.Imp.GetLotteryShareRecordByCode(shareCode)
		if getErr == nil && existing.ShareUserID == user.ID && existing.LotteryID == req.LotteryID {
			res.Code = 0
			res.Data = LotteryShareCreateResponse{
				ShareCode:   existing.ShareCode,
				ShareSource: existing.ShareSource,
				ShareStatus: existing.ShareStatus,
				Message:     "share_prepared",
			}
			writeJSON(w, res)
			return
		}
		res.Code = -1
		res.ErrorMsg = "Failed to create share record"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = LotteryShareCreateResponse{
		ShareCode:   shareCode,
		ShareSource: shareSource,
		ShareStatus: lotteryShareStatusPrepared,
		Message:     "share_prepared",
	}
	writeJSON(w, res)
}

// LotteryShareConfirmHandler 确认本次分享已真实发起
func LotteryShareConfirmHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	var req LotteryShareConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	req.LotteryID = strings.TrimSpace(req.LotteryID)
	req.ShareCode = strings.TrimSpace(req.ShareCode)
	if req.LotteryID == "" || req.ShareCode == "" {
		res.Code = -1
		res.ErrorMsg = "Lottery ID and share code are required"
		writeJSON(w, res)
		return
	}

	record, err := dao.Imp.GetLotteryShareRecordByCode(req.ShareCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = -1
		res.ErrorMsg = "Invalid share code"
		writeJSON(w, res)
		return
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to query share record"
		writeJSON(w, res)
		return
	}
	if record.LotteryID != req.LotteryID || record.ShareUserID != user.ID {
		res.Code = -1
		res.ErrorMsg = "Invalid share code"
		writeJSON(w, res)
		return
	}

	if err := markLotteryShareRecordShared(record); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to confirm share"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = LotteryShareConfirmResponse{
		ShareCode:   record.ShareCode,
		ShareSource: record.ShareSource,
		ShareStatus: record.ShareStatus,
		Message:     "share_confirmed",
	}
	writeJSON(w, res)
}

// LotteryShareRewardHandler 处理分享回流奖励
func LotteryShareRewardHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}

	receiver, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}

	var req LotteryShareRewardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	if req.LotteryID == "" || req.ShareUserID == "" {
		res.Code = -1
		res.ErrorMsg = "Lottery ID and share user ID are required"
		writeJSON(w, res)
		return
	}
	if strings.TrimSpace(req.ShareCode) == "" {
		res.Code = -1
		res.ErrorMsg = "Share code is required"
		writeJSON(w, res)
		return
	}

	if receiver.ID == req.ShareUserID {
		res.Code = 0
		res.Data = LotteryShareRewardResponse{
			Rewarded: false,
			Points:   0,
			Message:  "self_open_ignored",
		}
		writeJSON(w, res)
		return
	}

	if _, err := dao.Imp.GetLotteryByID(req.LotteryID); err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}

	if _, err := dao.Imp.GetUserByID(req.ShareUserID); err != nil {
		res.Code = -1
		res.ErrorMsg = "Share user not found"
		writeJSON(w, res)
		return
	}

	record, err := dao.Imp.GetLotteryShareRecordByCode(strings.TrimSpace(req.ShareCode))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = -1
		res.ErrorMsg = "Invalid share code"
		writeJSON(w, res)
		return
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to verify share code"
		writeJSON(w, res)
		return
	}
	if record.LotteryID != req.LotteryID || record.ShareUserID != req.ShareUserID {
		res.Code = -1
		res.ErrorMsg = "Invalid share code"
		writeJSON(w, res)
		return
	}
	if record.ShareStatus != lotteryShareStatusPrepared && record.ShareStatus != lotteryShareStatusShared {
		res.Code = -1
		res.ErrorMsg = "Invalid share status"
		writeJSON(w, res)
		return
	}

	if _, err := dao.Imp.GetLotteryShareReward(req.LotteryID, req.ShareUserID, receiver.ID); err == nil {
		res.Code = 0
		res.Data = LotteryShareRewardResponse{
			Rewarded: false,
			Points:   0,
			Message:  "already_rewarded",
		}
		writeJSON(w, res)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		res.Code = -1
		res.ErrorMsg = "Failed to query reward record"
		writeJSON(w, res)
		return
	}

	now := time.Now()
	updatedIntegral := 0
	err = db.Get().Transaction(func(tx *gorm.DB) error {
		reward := &model.LotteryShareReward{
			LotteryID:      req.LotteryID,
			ShareUserID:    req.ShareUserID,
			ReceiverUserID: receiver.ID,
			ShareCode:      req.ShareCode,
			RewardPoints:   10,
			Rewarded:       true,
			RewardedAt:     &now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(reward).Error; err != nil {
			return err
		}
		if reward.ID == 0 {
			return errShareRewardExists
		}

		result := tx.Model(&model.User{}).
			Where("id = ?", req.ShareUserID).
			Updates(map[string]interface{}{
				"integral":   gorm.Expr("integral + ?", 10),
				"updated_at": now,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		var updatedUser model.User
		if err := tx.Select("integral").Where("id = ?", req.ShareUserID).First(&updatedUser).Error; err != nil {
			return err
		}
		updatedIntegral = updatedUser.Integral
		return nil
	})
	if errors.Is(err, errShareRewardExists) {
		res.Code = 0
		res.Data = LotteryShareRewardResponse{
			Rewarded: false,
			Points:   0,
			Message:  "already_rewarded",
		}
		writeJSON(w, res)
		return
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to reward share"
		writeJSON(w, res)
		return
	}
	if err := markLotteryShareRecordShared(record); err != nil {
		// 回流奖励已完成，不因状态回写失败影响积分结果，只记录日志。
		w.Header().Add("X-Share-Confirm-Warn", "share_record_not_marked_shared")
	}

	res.Code = 0
	res.Data = LotteryShareRewardResponse{
		Rewarded:        true,
		Points:          10,
		Message:         "积分 +10",
		CurrentIntegral: updatedIntegral,
	}
	writeJSON(w, res)
}

func markLotteryShareRecordShared(record *model.LotteryShareRecord) error {
	if record == nil || record.ShareStatus == lotteryShareStatusShared {
		return nil
	}
	now := time.Now()
	record.ShareStatus = lotteryShareStatusShared
	record.SharedAt = &now
	record.UpdatedAt = now
	return dao.Imp.UpdateLotteryShareRecord(record)
}

func generateShareCode() string {
	return "share_" + strings.ReplaceAll(uuid.New().String(), "-", "")
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysqlDriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == 1062
}
