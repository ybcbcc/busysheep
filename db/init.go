package db

import (
	"fmt"
	"os"
	"time"
	"wxcloudrun-golang/db/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var dbInstance *gorm.DB

// Init 初始化数据库
func Init() error {

	source := "%s:%s@tcp(%s)/%s?readTimeout=1500ms&writeTimeout=1500ms&charset=utf8&loc=Local&&parseTime=true"
	user := os.Getenv("MYSQL_USERNAME")
	pwd := os.Getenv("MYSQL_PASSWORD")
	addr := os.Getenv("MYSQL_ADDRESS")
	dataBase := os.Getenv("MYSQL_DATABASE")
	if dataBase == "" {
		dataBase = "golang_demo"
	}
	source = fmt.Sprintf(source, user, pwd, addr, dataBase)
	fmt.Println("start init mysql with ", source)

	db, err := gorm.Open(mysql.Open(source), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // use singular table name, table for `User` would be `user` with this option enabled
		}})
	if err != nil {
		fmt.Println("DB Open error,err=", err.Error())
		return err
	}

	sqlDB, err := db.DB()
	if err != nil {
		fmt.Println("DB Init error,err=", err.Error())
		return err
	}

	// 用于设置连接池中空闲连接的最大数量
	sqlDB.SetMaxIdleConns(100)
	// 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(200)
	// 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(time.Hour)

	dbInstance = db

	// ==========================================
	// 数据库重置逻辑 (根据用户要求：清理旧数据并重建)
	// ==========================================
	
	// 1. 显式清理已知的残留表 (防止表名变更导致的残留)
	// 这里的表名是硬编码的，确保彻底清理
	legacyTables := []string{"user_lottery_record", "user", "lottery", "lottery_participant"}
	for _, tb := range legacyTables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", tb)).Error; err != nil {
			fmt.Printf("Warning: Failed to drop legacy table %s: %v\n", tb, err)
		}
	}

	// 2. 使用 GORM 的 DropTable 清理当前模型对应的表 (双重保险)
	if err := db.Migrator().DropTable(&model.User{}, &model.Lottery{}, &model.LotteryParticipant{}); err != nil {
		fmt.Printf("Drop tables failed: %v\n", err)
	}

	fmt.Println("Old tables and data cleared successfully.")
	
	// 3. 自动迁移创建新表
	// 注册新模型：User, Lottery, LotteryParticipant
	// 保留旧模型：CounterModel, Post
	err = db.AutoMigrate(&model.CounterModel{}, &model.User{}, &model.Lottery{}, &model.LotteryParticipant{}, &model.Post{})
	if err != nil {
		fmt.Println("DB Migrate error,err=", err.Error())
		return err
	}

	fmt.Println("finish init mysql with ", source)
	return nil
}

// Get ...
func Get() *gorm.DB {
	return dbInstance
}
