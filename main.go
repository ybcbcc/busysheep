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

	// Home
	http.HandleFunc("/api/home/list", service.HomeListHandler)

	// Lottery
	http.HandleFunc("/api/lottery/detail", service.LotteryDetailHandler)
	http.HandleFunc("/api/lottery/draw", service.LotteryDrawHandler)

	// Post
	http.HandleFunc("/api/upload", service.UploadHandler)
	http.HandleFunc("/api/post/create", service.CreatePostHandler)

	// User
	http.HandleFunc("/api/user/info", service.UserInfoHandler)
	http.HandleFunc("/api/user/lottery-history", service.UserLotteryHistoryHandler)
	http.HandleFunc("/api/user/publish-history", service.UserPublishHistoryHandler)
	http.HandleFunc("/api/member/info", service.MemberInfoHandler)

	// Static files for uploads
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	log.Fatal(http.ListenAndServe(":80", nil))
}
