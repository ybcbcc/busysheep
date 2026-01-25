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
	// Fallback for testing/template if env not set (DO NOT USE IN PROD)
	if appID == "" {
		// Mock mode or error
		// fmt.Println("Warning: WX_APP_ID not set")
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code", appID, appSecret, req.Code)
	
	// Mocking for template if no env vars, otherwise real call
    var openID string
    if appID == "" {
        // Mock OpenID for testing without real credentials
        openID = "mock-openid-" + req.Code
    } else {
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
			OpenID:    openID,
			Nickname:  "User_" + openID[:6],
			Points:    0,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	} else if err != nil {
		res.Code = -1
		res.ErrorMsg = "Database error"
		writeJSON(w, res)
		return
	}

	// Generate Token
	token := generateToken()
	user.Token = token
	user.UpdatedAt = time.Now()

	if err := dao.Imp.UpsertUser(user); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to save user"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = map[string]string{
		"token": token,
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
		// Try Authorization header
		token = r.Header.Get("Authorization")
	}
	if token == "" {
		return nil, fmt.Errorf("No token provided")
	}
	
	return dao.Imp.GetUserByToken(token)
}
