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

// InitTablesHandler 强制重置并初始化数据库表
func InitTablesHandler(w http.ResponseWriter, r *http.Request) {
	res := &JsonResult{}
	
	db := db.Get()
	results := make(map[string]string)

	// 1. 删除旧表
	tablesToDrop := []string{"user_lottery_record", "user", "lottery", "lottery_participant"}
	for _, tb := range tablesToDrop {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tb)).Error; err != nil {
			results[fmt.Sprintf("drop_%s", tb)] = fmt.Sprintf("Error: %v", err)
		} else {
			results[fmt.Sprintf("drop_%s", tb)] = "Success"
		}
	}
	
	// 2. 创建新表 SQL
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS user (
			id VARCHAR(32) PRIMARY KEY,
			open_id VARCHAR(64) NOT NULL,
			phone VARCHAR(11),
			phone_verified BOOLEAN DEFAULT FALSE,
			nickname VARCHAR(50) NOT NULL,
			avatar_url VARCHAR(500),
			integral INT DEFAULT 0,
			member_type ENUM('free', 'basic', 'advanced', 'annual') DEFAULT 'free',
			member_since DATETIME,
			member_expiry DATETIME,
			level INT DEFAULT 1,
			experience INT DEFAULT 0,
			device_fingerprint VARCHAR(64),
			invite_code VARCHAR(10),
			invited_by VARCHAR(32),
			status ENUM('active', 'frozen', 'banned') DEFAULT 'active',
			token VARCHAR(64),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE INDEX idx_open_id (open_id),
			UNIQUE INDEX idx_phone (phone),
			UNIQUE INDEX idx_invite_code (invite_code),
			INDEX idx_invited_by (invited_by),
			INDEX idx_token (token)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		
		`CREATE TABLE IF NOT EXISTS lottery (
			id VARCHAR(32) PRIMARY KEY,
			creator_id VARCHAR(32) NOT NULL,
			title VARCHAR(100) NOT NULL,
			image_url VARCHAR(255),
			description TEXT,
			prize_type ENUM('integral', 'membership', 'avatar_frame', 'chat_bubble', 'theme', 'external_vip') NOT NULL,
			prize_value INT NOT NULL,
			prize_name VARCHAR(100),
			cost_per_entry INT NOT NULL,
			max_participants INT NOT NULL,
			current_participants INT DEFAULT 0,
			win_probability DECIMAL(5,4) NOT NULL,
			win_count INT DEFAULT 1,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			actual_draw_time DATETIME,
			is_public BOOLEAN DEFAULT TRUE,
			member_only BOOLEAN DEFAULT FALSE,
			status ENUM('pending', 'active', 'finished', 'cancelled') DEFAULT 'pending',
			audit_status ENUM('pending', 'approved', 'rejected') DEFAULT 'pending',
			audit_reason VARCHAR(200),
			seed_string VARCHAR(500),
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_creator (creator_id),
			INDEX idx_status_endtime (status, end_time),
			INDEX idx_member_only (member_only)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
		
		`CREATE TABLE IF NOT EXISTS lottery_participant (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			lottery_id VARCHAR(32) NOT NULL,
			user_id VARCHAR(32) NOT NULL,
			integral_spent INT NOT NULL,
			entry_count INT DEFAULT 1,
			is_winner BOOLEAN DEFAULT FALSE,
			prize_received BOOLEAN DEFAULT FALSE,
			participated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE INDEX uk_lottery_user (lottery_id, user_id),
			INDEX idx_lottery (lottery_id),
			INDEX idx_user (user_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`,
	}
	
	
	for i, sql := range sqls {
		if err := db.Exec(sql).Error; err != nil {
			results[fmt.Sprintf("sql_%d", i)] = fmt.Sprintf("Error: %v", err)
		} else {
			results[fmt.Sprintf("sql_%d", i)] = "Success"
		}
	}
	
	res.Code = 0
	res.Data = results
	writeJSON(w, res)
}
