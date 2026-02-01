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
	envInfo := map[string]interface{}{
		"MYSQL_ADDRESS":  os.Getenv("MYSQL_ADDRESS"),
		"MYSQL_DATABASE_ENV": os.Getenv("MYSQL_DATABASE"),
		"MYSQL_USERNAME": os.Getenv("MYSQL_USERNAME"),
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

	// 4. 获取当前连接的数据库名称
	var currentDB string
	if err := db.Get().Raw("SELECT DATABASE()").Scan(&currentDB).Error; err != nil {
		envInfo["current_db_error"] = err.Error()
	} else {
		envInfo["current_db_connected"] = currentDB
	}

	// 5. 获取当前数据库的所有表
	var tables []string
	if err := db.Get().Raw("SHOW TABLES").Scan(&tables).Error; err != nil {
		envInfo["show_tables_error"] = err.Error()
	} else {
		envInfo["tables_in_current_db"] = tables
	}

	// 6. 获取服务器上的所有数据库 (如果权限允许)
	var allDBs []string
	if err := db.Get().Raw("SHOW DATABASES").Scan(&allDBs).Error; err != nil {
		envInfo["show_databases_error"] = err.Error()
	} else {
		envInfo["all_databases_on_server"] = allDBs
	}

	res.Code = 0
	res.Data = envInfo
	writeJSON(w, res)
}
