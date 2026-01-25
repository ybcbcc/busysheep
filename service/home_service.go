package service

import (
	"net/http"
	"sort"
	"time"
	"wxcloudrun-golang/db/dao"
)

// HomeItem 首页混合列表项
type HomeItem struct {
	Type      string      `json:"type"` // "lottery" or "post"
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"createdAt"`
}

// HomeListHandler 首页列表接口
func HomeListHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// 获取进行中的抽奖
	lotteries, err := dao.Imp.GetActiveLotteries()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch lotteries"
		writeJSON(w, res)
		return
	}

	// 获取已发布的帖子
	posts, err := dao.Imp.GetActivePosts()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch posts"
		writeJSON(w, res)
		return
	}

	// 混合并排序
	var items []HomeItem
	for _, l := range lotteries {
		items = append(items, HomeItem{Type: "lottery", Data: l, CreatedAt: l.CreatedAt})
	}
	for _, p := range posts {
		items = append(items, HomeItem{Type: "post", Data: p, CreatedAt: p.CreatedAt})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	res.Code = 0
	res.Data = items
	writeJSON(w, res)
}
