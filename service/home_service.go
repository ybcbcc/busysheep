package service

import (
	"net/http"
	"time"
	"wxcloudrun-golang/db/dao"
	"wxcloudrun-golang/db/model"
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

	// 不在服务端做时间比较或自动结算，交由前端按客户端当前时间判断

	// 现在的首页只展示 Lottery
	res.Code = 0
	bj := time.FixedZone("CST", 8*3600)
	out := make([]*model.Lottery, 0, len(lotteries))
	for _, l := range lotteries {
		c := *l
		c.StartTime = l.StartTime.In(bj)
		c.EndTime = l.EndTime.In(bj)
		out = append(out, &c)
	}
	res.Data = out
	writeJSON(w, res)
}
