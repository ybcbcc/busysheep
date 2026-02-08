package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// 中文注释：access_token 缓存结构，避免频繁请求微信接口
var (
	tokenMu     sync.Mutex
	cachedToken string
	expireAt    time.Time
)

// 中文注释：检测并返回缓存的 access_token，必要时主动刷新
func getAccessToken() (string, error) {
	tokenMu.Lock()
	defer tokenMu.Unlock()
	// 中文注释：提前1分钟刷新，避免边界时刻失效
	if cachedToken != "" && time.Now().Before(expireAt.Add(-time.Minute)) {
		return cachedToken, nil
	}
	getEnvMulti := func(names ...string) string {
		for _, n := range names {
			if v := os.Getenv(n); v != "" {
				return v
			}
		}
		return ""
	}
	appid := getEnvMulti("WX_APPID", "WX_APP_ID")
	secret := getEnvMulti("WX_APPSECRET", "WX_APP_SECRET")
	if appid == "" || secret == "" {
		return "", Err("缺少 WX_APPID/WX_APPSECRET 或 WX_APP_ID/WX_APP_SECRET 环境变量")
	}
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var data struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}
	if data.AccessToken == "" {
		return "", Errf("获取 access_token 失败: errcode=%d errmsg=%s", data.ErrCode, data.ErrMsg)
	}
	cachedToken = data.AccessToken
	ttl := data.ExpiresIn
	if ttl <= 0 {
		ttl = 7200
	}
	expireAt = time.Now().Add(time.Duration(ttl) * time.Second)
	return cachedToken, nil
}

// 中文注释：简易错误类型，便于返回错误信息
type simpleErr string

func (e simpleErr) Error() string { return string(e) }
func Err(msg string) error        { return simpleErr(msg) }
func Errf(format string, a ...interface{}) error {
	return simpleErr(fmt.Sprintf(format, a...))
}

// 中文注释：检查是否为配置检测请求（JSON/XML）
func isCheckContainerPath(contentType string, body []byte) bool {
	ct := strings.ToLower(contentType)
	if strings.Contains(ct, "application/json") {
		var obj struct {
			Action string `json:"action"`
		}
		if json.Unmarshal(body, &obj) == nil && strings.EqualFold(obj.Action, "CheckContainerPath") {
			return true
		}
	}
	if strings.Contains(ct, "xml") || bytes.Contains(bytes.ToLower(body), []byte("<xml>")) {
		// 中文注释：简单包含判断，避免复杂 XML 解析
		return bytes.Contains(bytes.ToLower(body), []byte("checkcontainerpath"))
	}
	return false
}

// 中文注释：客服消息推送处理入口
func KefuHandler(w http.ResponseWriter, r *http.Request) {
	// 中文注释：读取原始请求体
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	log.Printf("kefu push received: ct=%s openid=%s body_len=%d", r.Header.Get("Content-Type"), r.Header.Get("x-wx-openid"), len(body))

	// 中文注释：微信消息推送配置检测，需直接返回 success
	if isCheckContainerPath(r.Header.Get("Content-Type"), body) {
		w.Write([]byte("success"))
		return
	}

	// 中文注释：获取用户 openid（由微信侧在请求头传入）
	openid := r.Header.Get("x-wx-openid")
	if openid == "" {
		// 中文注释：没有 openid 时返回 success，避免错误
		w.Write([]byte("success"))
		return
	}

	// 中文注释：构造客服文本消息
	payload := map[string]interface{}{
		"touser":  openid,
		"msgtype": "text",
		"text": map[string]string{
			"content": "云托管接收消息推送成功，内容如下：\n" + string(body),
		},
	}
	sendOnce := func() (*http.Response, error) {
		accessToken, err := getAccessToken()
		if err != nil {
			return nil, err
		}
		url := "https://api.weixin.qq.com/cgi-bin/message/custom/send?access_token=" + accessToken
		buf, _ := json.Marshal(payload)
		req, _ := http.NewRequest("POST", url, bytes.NewReader(buf))
		req.Header.Set("Content-Type", "application/json")
		return http.DefaultClient.Do(req)
	}

	// 中文注释：发送一次，必要时因 token 失效重试一次
	resp, err := sendOnce()
	if err == nil && resp != nil {
		defer resp.Body.Close()
		var res struct {
			ErrCode int    `json:"errcode"`
			ErrMsg  string `json:"errmsg"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&res)
		log.Printf("kefu send result: errcode=%d errmsg=%s", res.ErrCode, res.ErrMsg)
		if res.ErrCode == 40001 || res.ErrCode == 41001 {
			// 中文注释：清空缓存后重试
			tokenMu.Lock()
			cachedToken = ""
			expireAt = time.Time{}
			tokenMu.Unlock()
			if resp2, err2 := sendOnce(); err2 == nil && resp2 != nil {
				defer resp2.Body.Close()
				var res2 struct {
					ErrCode int    `json:"errcode"`
					ErrMsg  string `json:"errmsg"`
				}
				_ = json.NewDecoder(resp2.Body).Decode(&res2)
				log.Printf("kefu resend result: errcode=%d errmsg=%s", res2.ErrCode, res2.ErrMsg)
			}
		} else if err != nil {
			log.Printf("kefu send error: %v", err)
		}
	} else if err != nil {
		log.Printf("kefu sendOnce error: %v", err)
	}

	// 中文注释：按照微信侧要求返回 success
	w.Write([]byte("success"))
}
