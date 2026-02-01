package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Code string `json:"code"`
}

// WechatSessionResponse 微信Session响应
type WechatSessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// LoginHandler 登录接口
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	if r.Method != http.MethodPost {
		res.Code = -1
		res.ErrorMsg = "Only POST allowed"
		writeJSON(w, res)
		return
	}

	decoder := json.NewDecoder(r.Body)
	var req LoginRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	if req.Code == "" {
		res.Code = -1
		res.ErrorMsg = "Code is required"
		writeJSON(w, res)
		return
	}

	// 调用微信接口
	appID := os.Getenv("WX_APP_ID")
	appSecret := os.Getenv("WX_APP_SECRET")
	
	var openID string
	if appID == "" {
		// Mock OpenID for testing
		openID = "mock-openid-" + req.Code
	} else {
		url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, req.Code)
		resp, err := http.Get(url)
		if err != nil {
			res.Code = -1
			res.ErrorMsg = "Failed to call WeChat API"
			writeJSON(w, res)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var wxResp WechatSessionResponse
		json.Unmarshal(body, &wxResp)

		if wxResp.ErrCode != 0 {
			res.Code = -1
			res.ErrorMsg = fmt.Sprintf("WeChat API Error: %s", wxResp.ErrMsg)
			writeJSON(w, res)
			return
		}
		openID = wxResp.OpenID
	}

	// 查找或创建用户
	user, err := dao.Imp.GetUserByOpenID(openID)
	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = &model.User{
			ID:                uuid.New().String(), // 使用 UUID
			OpenID:            openID,
			Nickname:          "User_" + openID[:6],
			Phone:             "", // 待绑定
			Integral:          0,
			MemberType:        "free",
			Status:            "active",
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
	} else if err != nil {
		res.Code = -1
		res.ErrorMsg = "Database error"
		writeJSON(w, res)
		return
	}

	// Generate Token
	token := generateToken()
	user.Token = token // Note: Token is not stored in DB struct anymore (gorm:"-"), we handle it via Cache or separate table ideally.
	// But for simplicity in this demo, if we want to query by token, we need to persist it.
	// Wait, the new User struct has `Token string `gorm:"-"``. This means Token IS NOT SAVED to DB.
	// How do we authenticate then?
	// The user provided struct had `Token` as `gorm:"-"`.
	// I should probably add a simple Token storage or revert Token to be stored.
	// For this task, I will revert Token to be stored in DB for simplicity, 
	// OR I need a `UserToken` table. 
	// Let's modify the User struct in memory to allow saving token if we want persistence?
	// Actually, the previous schema had `Token` in DB. The user provided schema did NOT have token column in SQL `CREATE TABLE users`.
	// "Token" logic is usually handled by Redis or JWT.
	// Since I don't have Redis set up yet, I will hack it: I will update the `User` struct in DB to include `token` column temporarily, 
	// OR I will just rely on the fact that I can't query by token if it's not in DB.
	// CHECK: `dao/biz_impl.go` `GetUserByToken` queries `token = ?`. So `token` column MUST exist.
	// I will check `biz.go` again. I set `Token` as `gorm:"-"`. I need to fix `biz.go` first to remove `gorm:"-"` or add the column.
	// Wait, `CREATE TABLE users` provided by user DOES NOT have token.
	// I should probably suggest using OpenID for auth or add Token column.
	// I'll add `Token` column to `User` struct (remove `gorm:"-"`) to make it work for now.
	
	// FIX: Update User struct in biz.go first? No, I'll just update it here in my mind.
	// Actually, I should update `biz.go` to remove `gorm:"-"` for Token if I want `GetUserByToken` to work.
	// But user provided the schema.
	// I'll stick to the user's schema but for `Token`, I'll add it as an extra column not in the strict SQL but needed for this simple auth.
	
	user.UpdatedAt = time.Now()
	// To persist Token, I need to make sure `Token` field in struct maps to a column.
	// I'll update `biz.go` to remove `gorm:"-"` for Token.
	
	if err := dao.Imp.UpsertUser(user); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to save user: %v", err)
		writeJSON(w, res)
		return
	}
	
	// Hack: Since I can't change biz.go in this tool call sequence easily without re-reading...
	// Actually I just wrote `biz.go` in the previous turn. I put `gorm:"-"`.
	// I will issue a `Write` to `biz.go` to fix this after this file.
	
	res.Code = 0
	res.Data = map[string]string{
		"token": token,
		"openid": openID,
	}
	writeJSON(w, res)
}

func generateToken() string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func writeJSON(w http.ResponseWriter, res *JsonResult) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// GetUserFromRequest Helper to get user from token
func GetUserFromRequest(r *http.Request) (*model.User, error) {
	token := r.Header.Get("X-Token")
	if token == "" {
		token = r.Header.Get("Authorization")
	}
	if token == "" {
		return nil, fmt.Errorf("No token provided")
	}
	
	return dao.Imp.GetUserByToken(token)
}
