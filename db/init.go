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
	// 数据库初始化逻辑
	// 注意：已移除 DropTable 逻辑，防止数据丢失。
	// ==========================================
	
	// 自动迁移创建新表 (只会新增表或列，不会删除数据)
	// 注册新模型：User, Lottery, LotteryParticipant
	// 保留旧模型：CounterModel
	// 注意：Post 模型不再自动迁移，如果需要清理旧数据，请手动操作数据库
	err = db.AutoMigrate(&model.CounterModel{}, &model.User{}, &model.Lottery{}, &model.LotteryParticipant{}, &model.Activity{}, &model.ActivityExposure{})
	if err != nil {
		fmt.Println("DB Migrate error,err=", err.Error())
		return err
	}

	// 额外迁移：删除历史唯一索引（如果存在）
	dropIndexes := []string{
		"ALTER TABLE user DROP INDEX idx_open_id",
		"ALTER TABLE user DROP INDEX idx_phone",
		"ALTER TABLE user DROP INDEX idx_invite_code",
		"ALTER TABLE lottery_participant DROP INDEX uk_lottery_user",
	}
	for _, sql := range dropIndexes {
		if err := db.Exec(sql).Error; err != nil {
			fmt.Println("Drop index ignore error:", sql, "err=", err.Error())
		}
	}

	fmt.Println("finish init mysql with ", source)
	return nil
}

// Get ...
func Get() *gorm.DB {
	return dbInstance
}
