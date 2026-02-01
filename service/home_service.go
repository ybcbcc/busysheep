package service

import (
	"net/http"
	"wxcloudrun-golang/db/dao"
)

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

	// 现在的首页只展示 Lottery
	res.Code = 0
	res.Data = lotteries
	writeJSON(w, res)
}
