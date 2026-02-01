package model

import "time"

// User 用户模型
// 根据新需求重建，对应 CREATE TABLE users
type User struct {
	ID                string     `gorm:"primaryKey;type:varchar(32)" json:"id"` // UUID
	OpenID            string     `gorm:"uniqueIndex:idx_open_id;type:varchar(64);not null" json:"openId"`
	Phone             *string    `gorm:"uniqueIndex:idx_phone;type:varchar(11)" json:"phone"`
	PhoneVerified     bool       `gorm:"default:false" json:"phoneVerified"`
	Nickname          string     `gorm:"type:varchar(50);not null" json:"nickname"`
	AvatarURL         string     `gorm:"type:varchar(500);column:avatar_url" json:"avatarUrl"`
	Integral          int        `gorm:"default:0" json:"integral"`
	MemberType        string     `gorm:"type:enum('free', 'basic', 'advanced', 'annual');default:'free'" json:"memberType"`
	MemberSince       *time.Time `gorm:"type:datetime" json:"memberSince"`
	MemberExpiry      *time.Time `gorm:"type:datetime" json:"memberExpiry"`
	Level             int        `gorm:"default:1" json:"level"`
	Experience        int        `gorm:"default:0" json:"experience"`
	DeviceFingerprint string     `gorm:"type:varchar(64)" json:"deviceFingerprint"`
	InviteCode        string     `gorm:"uniqueIndex:idx_invite_code;type:varchar(10)" json:"inviteCode"`
	InvitedBy         string     `gorm:"index:idx_invited_by;type:varchar(32)" json:"invitedBy"`
	Status            string     `gorm:"type:enum('active', 'frozen', 'banned');default:'active'" json:"status"`
	Token             string     `gorm:"type:varchar(64);index" json:"token"` // 存储用户token用于认证
	CreatedAt         time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt         time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// Lottery 抽奖活动模型
// 对应 CREATE TABLE lotteries
type Lottery struct {
	ID                  string     `gorm:"primaryKey;type:varchar(32)" json:"id"`
	CreatorID           string     `gorm:"index:idx_creator;type:varchar(32);not null" json:"creatorId"`
	Title               string     `gorm:"type:varchar(100);not null" json:"title"`
	ImageURL            string     `gorm:"type:varchar(255)" json:"imageUrl"` // 活动封面图
	Description         string     `gorm:"type:text" json:"description"`
	PrizeType           string     `gorm:"type:enum('integral', 'membership', 'avatar_frame', 'chat_bubble', 'theme', 'external_vip');not null" json:"prizeType"`
	PrizeValue          int        `gorm:"not null" json:"prizeValue"`
	PrizeName           string     `gorm:"type:varchar(100)" json:"prizeName"`
	CostPerEntry        int        `gorm:"not null" json:"costPerEntry"`
	MaxParticipants     int        `gorm:"not null" json:"maxParticipants"`
	CurrentParticipants int        `gorm:"default:0" json:"currentParticipants"`
	WinProbability      float64    `gorm:"type:decimal(5,4);not null" json:"winProbability"` // 0.0001-1.0000
	WinCount            int        `gorm:"default:1" json:"winCount"`
	StartTime           time.Time  `gorm:"type:datetime;not null" json:"startTime"`
	EndTime             time.Time  `gorm:"type:datetime;not null;index:idx_status_endtime" json:"endTime"`
	ActualDrawTime      *time.Time `gorm:"type:datetime" json:"actualDrawTime"`
	IsPublic            bool       `gorm:"default:true" json:"isPublic"`
	MemberOnly          bool       `gorm:"default:false;index:idx_member_only" json:"memberOnly"`
	Status              string     `gorm:"type:enum('pending', 'active', 'finished', 'cancelled');default:'pending';index:idx_status_endtime;index:idx_member_only" json:"status"`
	AuditStatus         string     `gorm:"type:enum('pending', 'approved', 'rejected');default:'pending'" json:"auditStatus"`
	AuditReason         string     `gorm:"type:varchar(200)" json:"auditReason"`
	SeedString          string     `gorm:"type:varchar(500)" json:"seedString"` // 开奖随机种子
	CreatedAt           time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt           time.Time  `gorm:"type:datetime;default:CURRENT_TIMESTAMP" json:"updatedAt"`
}

// LotteryParticipant 抽奖参与记录
// 对应 CREATE TABLE lottery_participants
type LotteryParticipant struct {
	ID             int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	LotteryID      string    `gorm:"type:varchar(32);uniqueIndex:uk_lottery_user;index:idx_lottery;not null" json:"lotteryId"`
	UserID         string    `gorm:"type:varchar(32);uniqueIndex:uk_lottery_user;index:idx_user;not null" json:"userId"`
	IntegralSpent  int       `gorm:"not null" json:"integralSpent"`
	EntryCount     int       `gorm:"default:1" json:"entryCount"`
	IsWinner       bool      `gorm:"default:false;index:idx_lottery" json:"isWinner"`
	PrizeReceived  bool      `gorm:"default:false" json:"prizeReceived"`
	ParticipatedAt time.Time `gorm:"type:datetime;default:CURRENT_TIMESTAMP;index:idx_user" json:"participatedAt"`
}

// Post 发布内容模型 (保留原业务逻辑，适配新ID类型)
type Post struct {
	ID        int32     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"index;type:varchar(32)" json:"userId"` // 适配 User.ID 类型变化
	Content   string    `gorm:"type:text" json:"content"`
	ImageURL  string    `gorm:"type:varchar(255)" json:"imageUrl"`
	Location  string    `gorm:"type:varchar(128)" json:"location"`
	Status    int       `gorm:"default:0" json:"status"` // 0: 待审核, 1: 已发布
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}
