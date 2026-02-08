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
	"gorm.io/gorm"
)

type CreateActivityRequest struct {
	Name                string `json:"name"`
	ImageURL            string `json:"imageUrl"`
	Content             string `json:"content"`
	StartTime           string `json:"startTime"`       // "2006-01-02 15:04:05"
	DurationMinutes     int    `json:"durationMinutes"` // 单位：分钟
	AppearanceFrequency int    `json:"appearanceFrequency"`
}

func CreateActivityHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	user, err := GetUserFromRequest(r)
	if err != nil {
		res.Code = 401
		res.ErrorMsg = "Unauthorized"
		writeJSON(w, res)
		return
	}
	if user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req CreateActivityRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	start := time.Now()
	if req.StartTime != "" {
		if t, err := time.Parse("2006-01-02 15:04:05", req.StartTime); err == nil {
			start = t
		}
	}
	duration := 0
	if req.DurationMinutes > 0 {
		duration = req.DurationMinutes
	}

	act := &model.Activity{
		ID:                  strings.ReplaceAll(uuid.New().String(), "-", ""),
		CreatorID:           user.ID,
		Name:                req.Name,
		ImageURL:            req.ImageURL,
		Content:             req.Content,
		StartTime:           start,
		DurationMinutes:     duration,
		AppearanceFrequency: req.AppearanceFrequency,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := dao.Imp.CreateActivity(act); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to create activity: %v", err)
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = act
	writeJSON(w, res)
}

func ActivityLatestHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	act, err := dao.Imp.GetLatestActivity()
	if err == gorm.ErrRecordNotFound {
		res.Code = 0
		res.Data = nil
		writeJSON(w, res)
		return
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch latest activity"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = act
	writeJSON(w, res)
}

func ActivityCurrentHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	now := time.Now().Unix()
	act, err := dao.Imp.GetCurrentActiveActivity(now)
	if err == gorm.ErrRecordNotFound {
		res.Code = 0
		res.Data = nil
		writeJSON(w, res)
		return
	}
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch current activity"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = act
	writeJSON(w, res)
}

type ActivityListItem struct {
	ID                  string    `json:"id"`
	Name                string    `json:"name"`
	ImageURL            string    `json:"imageUrl"`
	Content             string    `json:"content"`
	StartTime           time.Time `json:"startTime"`
	DurationMinutes     int       `json:"durationMinutes"`
	AppearanceFrequency int       `json:"appearanceFrequency"`
	Status              string    `json:"status"`
}

func ActivityListHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}
	acts, err := dao.Imp.GetAllActivities()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch activities"
		writeJSON(w, res)
		return
	}
	now := time.Now()
	list := make([]ActivityListItem, 0, len(acts))
	for _, a := range acts {
		end := a.StartTime.Add(time.Duration(a.DurationMinutes) * time.Minute)
		status := "active"
		if now.Before(a.StartTime) {
			status = "not_started"
		} else if !now.Before(end) {
			status = "finished"
		}
		list = append(list, ActivityListItem{
			ID:                  a.ID,
			Name:                a.Name,
			ImageURL:            a.ImageURL,
			Content:             a.Content,
			StartTime:           a.StartTime,
			DurationMinutes:     a.DurationMinutes,
			AppearanceFrequency: a.AppearanceFrequency,
			Status:              status,
		})
	}
	res.Code = 0
	res.Data = list
	writeJSON(w, res)
}
