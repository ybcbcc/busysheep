package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"strings"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ActivityCreateRequest struct {
	Name            string    `json:"name"`
	ImageURL        string    `json:"imageUrl"`
	Content         string    `json:"content"`
	StartTime       time.Time `json:"startTime"`
	DurationMinutes int       `json:"durationMinutes"`
	AppearanceCount int       `json:"appearanceCount"`
}

type ActivityUpdateRequest struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	ImageURL        string    `json:"imageUrl"`
	Content         string    `json:"content"`
	StartTime       time.Time `json:"startTime"`
	DurationMinutes int       `json:"durationMinutes"`
	AppearanceCount int       `json:"appearanceCount"`
}

func ActivityListHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, _ := GetUserFromRequest(r)
	var list []*model.Activity
	var err error
	if user != nil && user.Role == "admin" {
		list, err = dao.Imp.GetAllActivities()
	} else {
		list, err = dao.Imp.GetActiveActivities()
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch activities"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = list
	writeJSON(w, res)
}

func ActivityDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	id := r.URL.Query().Get("id")
	if id == "" {
		res.Code = -1
		res.ErrorMsg = "id is required"
		writeJSON(w, res)
		return
	}
	activity, err := dao.Imp.GetActivityByID(id)
	if err == gorm.ErrRecordNotFound {
		res.Code = 0
		res.Data = nil
		writeJSON(w, res)
		return
	} else if err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("DB error: %v", err)
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = activity
	writeJSON(w, res)
}

func ActivityCreateHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = -1
		res.ErrorMsg = "No permission"
		writeJSON(w, res)
		return
	}
	var req ActivityCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}
	now := time.Now()
	activity := &model.Activity{
		ID:              uuidTo32(uuid.New().String()),
		CreatorID:       user.ID,
		Name:            req.Name,
		ImageURL:        req.ImageURL,
		Content:         req.Content,
		StartTime:       req.StartTime,
		DurationMinutes: req.DurationMinutes,
		AppearanceCount: req.AppearanceCount,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := dao.Imp.CreateActivity(activity); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to create activity: %v", err)
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = activity
	writeJSON(w, res)
}

func ActivityUpdateHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = -1
		res.ErrorMsg = "No permission"
		writeJSON(w, res)
		return
	}
	var req ActivityUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}
	activity, err := dao.Imp.GetActivityByID(req.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Activity not found"
		writeJSON(w, res)
		return
	}
	activity.Name = req.Name
	activity.ImageURL = req.ImageURL
	activity.Content = req.Content
	activity.StartTime = req.StartTime
	activity.DurationMinutes = req.DurationMinutes
	activity.AppearanceCount = req.AppearanceCount
	activity.UpdatedAt = time.Now()
	if err := dao.Imp.UpdateActivity(activity); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to update activity: %v", err)
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = activity
	writeJSON(w, res)
}

func ActivityDeleteHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = -1
		res.ErrorMsg = "No permission"
		writeJSON(w, res)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		res.Code = -1
		res.ErrorMsg = "id is required"
		writeJSON(w, res)
		return
	}
	if err := dao.Imp.DeleteActivity(id); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to delete activity: %v", err)
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = "ok"
	writeJSON(w, res)
}

func ActivityAdLatestHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	activity, err := dao.Imp.GetLatestActiveActivity()
	if err == gorm.ErrRecordNotFound {
		res.Code = 0
		res.Data = nil
		writeJSON(w, res)
		return
	} else if err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("DB error: %v", err)
		writeJSON(w, res)
		return
	}
	if activity.AppearanceCount > 0 {
		activity.AppearanceCount = activity.AppearanceCount - 1
		activity.UpdatedAt = time.Now()
		_ = dao.Imp.UpdateActivity(activity)
	}
	res.Code = 0
	res.Data = activity
	writeJSON(w, res)
}

func ActivityAnnouncementsHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}
	list, err := dao.Imp.GetRecentActivities(limit)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch announcements"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = list
	writeJSON(w, res)
}

func uuidTo32(s string) string {
	return strings.ReplaceAll(s, "-", "")
}
