package service

import (
	"net/http"
	"wxcloudrun-golang/db/dao"
	"time"
)

// HomeListHandler 首页列表接口
func HomeListHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}

	// 获取全部抽奖（含进行中与已结束）
	lotteries, err := dao.Imp.GetAllLotteries()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = "Failed to fetch lotteries"
		writeJSON(w, res)
		return
	}

	// 到期自动结算：为了首页展示正确，将已到结束时间但未结算的活动统一置为已结束
	nowBJ := time.Now().Add(8 * time.Hour)
	for _, l := range lotteries {
		if l.Status != "finished" && nowBJ.After(l.EndTime) {
			finalizeRemainingPrizes(l)
		}
	}

	// 现在的首页只展示 Lottery
	res.Code = 0
	res.Data = lotteries
	writeJSON(w, res)
}
