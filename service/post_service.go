package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// UploadHandler 文件上传接口
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// Limit 10MB
	r.ParseMultipartForm(10 << 20)

	file, header, err := r.FormFile("file")
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Retrieve file failed"
		writeJSON(w, res)
		return
	}
	defer file.Close()

	// 检查是否配置了COS环境变量
	cosBucket := os.Getenv("COS_BUCKET")
	cosRegion := os.Getenv("COS_REGION")
	// 显式获取密钥 (修复 403 AccessDenied 问题)
	cosSecretID := os.Getenv("COS_SECRET_ID")
	cosSecretKey := os.Getenv("COS_SECRET_KEY")

	// 生成唯一文件名
	filename := fmt.Sprintf("uploads/%d_%s", time.Now().UnixNano(), header.Filename)
	var fileURL string

	if cosBucket != "" && cosRegion != "" {
		// 使用 COS 上传
		u, _ := url.Parse(fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cosBucket, cosRegion))
		b := &cos.BaseURL{BucketURL: u}
		
		// 优先使用显式密钥，否则降级到自动注入(可能权限不足)
		var client *cos.Client
		if cosSecretID != "" && cosSecretKey != "" {
			client = cos.NewClient(b, &http.Client{
				Transport: &cos.AuthorizationTransport{
					SecretID:  cosSecretID,
					SecretKey: cosSecretKey,
				},
			})
		} else {
			client = cos.NewClient(b, &http.Client{
				Transport: &cos.AuthorizationTransport{
					// 微信云托管容器内部会自动注入临时密钥
				},
			})
		}

		_, err = client.Object.Put(context.Background(), filename, file, nil)
		if err != nil {
			res.Code = -1
			res.ErrorMsg = fmt.Sprintf("Upload to COS failed: %v", err)
			writeJSON(w, res)
			return
		}
		// 返回 COS 公网访问 URL
		fileURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", cosBucket, cosRegion, filename)
	} else {
		// 降级到本地存储 (Local Storage Fallback)
		uploadDir := "./uploads"
		if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
			os.Mkdir(uploadDir, 0755)
		}
		
		localPath := filepath.Join(uploadDir, filepath.Base(filename))
		f, err := os.Create(localPath)
		if err != nil {
			res.Code = -1
			res.ErrorMsg = "Save file failed"
			writeJSON(w, res)
			return
		}
		defer f.Close()
		io.Copy(f, file)
		
		fileURL = fmt.Sprintf("/uploads/%s", filepath.Base(filename))
	}

	res.Code = 0
	res.Data = map[string]string{
		"url": fileURL,
	}
	writeJSON(w, res)
}

// CreatePostRequest 发布请求
type CreatePostRequest struct {
	Content  string `json:"content"`
	ImageURL string `json:"imageUrl"`
	Location string `json:"location"`
}

// CreatePostHandler 发布帖子接口
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
	var req CreatePostRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// 3. Save
	post := &model.Post{
		UserID:    user.ID,
		Content:   req.Content,
		ImageURL:  req.ImageURL,
		Location:  req.Location,
		Status:    1, // 暂时跳过审核，直接发布 (1: Published)
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := dao.Imp.CreatePost(post); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to create post"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	res.Data = post
	writeJSON(w, res)
}

// UpdatePostRequest 更新请求
type UpdatePostRequest struct {
	ID       int32  `json:"id"`
	Content  string `json:"content"`
	ImageURL string `json:"imageUrl"`
}

// UpdatePostHandler 更新帖子接口
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
	var req UpdatePostRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// 3. Check ownership
	post, err := dao.Imp.GetPostByID(req.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Post not found"
		writeJSON(w, res)
		return
	}

	if post.UserID != user.ID {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}

	// 4. Update
	post.Content = req.Content
	post.ImageURL = req.ImageURL
	post.UpdatedAt = time.Now()
	// post.Status = 0 // 可选：修改后重新审核

	if err := dao.Imp.UpdatePost(post); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to update post"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	writeJSON(w, res)
}

// DeletePostRequest 删除请求
type DeletePostRequest struct {
	ID int32 `json:"id"`
}

// DeletePostHandler 删除帖子接口
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

	// 2. Parse
	decoder := json.NewDecoder(r.Body)
	var req DeletePostRequest
	if err := decoder.Decode(&req); err != nil {
		res.Code = -1
		res.ErrorMsg = "Invalid JSON"
		writeJSON(w, res)
		return
	}

	// 3. Check ownership
	post, err := dao.Imp.GetPostByID(req.ID)
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Post not found"
		writeJSON(w, res)
		return
	}

	if post.UserID != user.ID {
		res.Code = 403
		res.ErrorMsg = "Forbidden"
		writeJSON(w, res)
		return
	}

	// 4. Delete
	if err := dao.Imp.DeletePost(req.ID); err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to delete post"
		writeJSON(w, res)
		return
	}

	res.Code = 0
	writeJSON(w, res)
}
