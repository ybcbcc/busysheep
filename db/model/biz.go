package model

import "time"

// User 用户模型
type User struct {
	ID             int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	OpenID         string    `gorm:"uniqueIndex;type:varchar(64)" json:"openid"`
	Nickname       string    `gorm:"type:varchar(64)" json:"nickname"`
	AvatarURL      string    `gorm:"type:varchar(255)" json:"avatarUrl"`
	Points         int       `gorm:"default:0" json:"points"`
	IsMember       bool      `gorm:"default:false" json:"isMember"`
	MemberExpireAt *time.Time `json:"memberExpireAt"`
	Token          string    `gorm:"index;type:varchar(64)" json:"token"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// Lottery 抽奖活动模型
type Lottery struct {
	ID              int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	Title           string    `gorm:"type:varchar(128)" json:"title"`
	Description     string    `gorm:"type:text" json:"description"`
	Status          int       `gorm:"default:1" json:"status"` // 1: 进行中, 0: 结束
	Probability     float64   `gorm:"type:decimal(5,4)" json:"probability"`
	PrizeType       string    `gorm:"type:varchar(32);default:'virtual'" json:"prizeType"` // virtual, real
	Cost            int       `gorm:"default:0" json:"cost"`
	MaxParticipants int       `gorm:"default:0" json:"maxParticipants"`
	DrawTime        time.Time `json:"drawTime"`
	UserID          int32     `gorm:"index" json:"userId"` // 新增：关联创建者ID
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UserLotteryRecord 抽奖记录
type UserLotteryRecord struct {
	ID        int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int32     `gorm:"index" json:"userId"`
	LotteryID int32     `gorm:"index" json:"lotteryId"`
	PrizeName string    `gorm:"type:varchar(128)" json:"prizeName"`
	CreatedAt time.Time `json:"createdAt"`
}

// Post 发布内容模型
type Post struct {
	ID        int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int32     `gorm:"index" json:"userId"`
	Content   string    `gorm:"type:text" json:"content"`
	ImageURL  string    `gorm:"type:varchar(255)" json:"imageUrl"`
	Location  string    `gorm:"type:varchar(128)" json:"location"`
	Status    int       `gorm:"default:0" json:"status"` // 0: 待审核, 1: 已发布
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
