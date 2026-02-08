package main

import (
	"fmt"
	"log"
	"net/http"
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/service"
)

func main() {
	if err := db.Init(); err != nil {
		panic(fmt.Sprintf("mysql init failed with %+v", err))
	}

	// Original routes
	http.HandleFunc("/", service.IndexHandler)
	http.HandleFunc("/api/count", service.CounterHandler)

	// New routes
	// Auth
	http.HandleFunc("/api/auth/login", service.LoginHandler)
	
	// Debug
	http.HandleFunc("/api/debug/db-check", service.DBCheckHandler)
	http.HandleFunc("/api/debug/init-tables", service.InitTablesHandler)
	http.HandleFunc("/api/debug/clear-all", service.ClearAllHandler)
	http.HandleFunc("/api/debug/clear-table", service.ClearTableHandler)

	// Home
	http.HandleFunc("/api/home/list", service.HomeListHandler)

	// Lottery
	http.HandleFunc("/api/lottery/detail", service.LotteryDetailHandler)
	http.HandleFunc("/api/lottery/draw", service.LotteryDrawHandler)

	// Admin
	http.HandleFunc("/api/admin/lotteries", service.AdminLotteryListHandler)
	http.HandleFunc("/api/admin/lottery/detail", service.AdminLotteryDetailHandler)
	http.HandleFunc("/api/admin/lottery/audit", service.AdminLotteryAuditHandler)

	// Post (Upload logic moved to frontend wx.cloud.uploadFile)
	http.HandleFunc("/api/post/create", service.CreatePostHandler)
	http.HandleFunc("/api/post/update", service.UpdatePostHandler)
	http.HandleFunc("/api/post/delete", service.DeletePostHandler)

	// Activity
	http.HandleFunc("/api/activity/create", service.CreateActivityHandler)
	http.HandleFunc("/api/activity/latest", service.ActivityLatestHandler)
	http.HandleFunc("/api/activity/current", service.ActivityCurrentHandler)
	http.HandleFunc("/api/activity/list", service.ActivityListHandler)

	// User
	http.HandleFunc("/api/user/info", service.UserInfoHandler)
	http.HandleFunc("/api/user/update", service.UpdateUserHandler)
	http.HandleFunc("/api/user/lottery-history", service.UserLotteryHistoryHandler)
	http.HandleFunc("/api/user/publish-history", service.UserPublishHistoryHandler)
	http.HandleFunc("/api/member/info", service.MemberInfoHandler)

	// Static files for uploads
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// 微信客服消息推送路径
	http.HandleFunc("/foo/bar", service.KefuHandler)

	log.Fatal(http.ListenAndServe(":80", nil))
}
