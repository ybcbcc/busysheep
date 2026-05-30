package service

import (
	"encoding/json"
	"net/http"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
	"time"
)

type AdminAuditListRequest struct {
	AuditStatus string `json:"auditStatus"`
}

func AdminLotteryListHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}
	status := r.URL.Query().Get("auditStatus")
	if status == "" {
		status = "pending"
	}
	var list []*model.Lottery
	lotteries, err := func() ([]*model.Lottery, error) {
		if status == "all" {
			return dao.Imp.GetAllLotteries()
		}
		return dao.Imp.GetLotteriesByAuditStatus(status)
	}()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch lotteries"
		writeJSON(w, res)
		return
	}
	list = lotteries
	res.Code = 0
	res.Data = list
	writeJSON(w, res)
}

func AdminLotteryDetailHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}
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

type AdminAuditRequest struct {
	ID     string `json:"id"`
	Action string `json:"action"`
	Reason string `json:"reason"`
}

func AdminLotteryAuditHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}
	decoder := json.NewDecoder(r.Body)
	var req AdminAuditRequest
	if err := decoder.Decode(&req); err != nil || req.ID == "" {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}
	lottery, err := dao.Imp.GetLotteryByID(req.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Lottery not found"
		writeJSON(w, res)
		return
	}
	bj := time.FixedZone("CST", 8*3600)
	nowBJ := time.Now().In(bj)
	startBJ := lottery.StartTime.In(bj)
	if nowBJ.After(startBJ) {
		res.Code = -1
		res.ErrorMsg = "Cannot audit after start"
		writeJSON(w, res)
		return
	}
	if req.Action == "approved" {
		lottery.AuditStatus = "approved"
		lottery.Status = "active"
		lottery.AuditReason = ""
	} else if req.Action == "rejected" {
		lottery.AuditStatus = "rejected"
		lottery.AuditReason = req.Reason
	} else {
		res.Code = -1
		res.ErrorMsg = "Invalid action"
		writeJSON(w, res)
		return
	}
	if err := dao.Imp.UpdateLottery(lottery); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to update lottery"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.Data = map[string]string{"result": "ok"}
	writeJSON(w, res)
}

func AdminLotteryDeleteHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	user, err := GetUserFromRequest(r)
	if err != nil || user.Role != "admin" {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		decoder := json.NewDecoder(r.Body)
		var req struct{ ID string `json:"id"` }
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
	if err := dao.Imp.DeleteLottery(id); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to delete lottery"
		writeJSON(w, res)
		return
	}
	res.Code = 0
	res.ErrorMsg = "Success"
	writeJSON(w, res)
}
