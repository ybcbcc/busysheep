package service

import (
	"fmt"
	"net/http"
	"os"
	"wxcloudrun-golang/db"
)

// DBCheckHandler 数据库连接检查接口
func DBCheckHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	
	// 1. 检查环境变量
	envInfo := map[string]string{
		"MYSQL_ADDRESS":  os.Getenv("MYSQL_ADDRESS"),
		"MYSQL_DATABASE": os.Getenv("MYSQL_DATABASE"),
		"MYSQL_USERNAME": os.Getenv("MYSQL_USERNAME"),
		// 密码不返回
	}

	// 2. 检查数据库连接对象
	sqlDB, err := db.Get().DB()
	if err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to get generic database object: %v", err)
		res.Data = envInfo
		writeJSON(w, res)
		return
	}

	// 3. Ping 数据库
	if err := sqlDB.Ping(); err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Database Ping failed: %v", err)
		res.Data = envInfo
		writeJSON(w, res)
		return
	}

	// 4. 检查表是否存在
	var tableCount int64
	// 检查 user 表是否存在
	if err := db.Get().Raw("SELECT count(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'user'").Scan(&tableCount).Error; err != nil {
		res.Code = -1
		res.ErrorMsg = fmt.Sprintf("Failed to query information_schema: %v", err)
		writeJSON(w, res)
		return
	}

	envInfo["user_table_exists"] = fmt.Sprintf("%v", tableCount > 0)

	res.Code = 0
	res.Data = envInfo
	writeJSON(w, res)
}
