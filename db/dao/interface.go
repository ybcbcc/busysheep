package dao

import (
	"wxcloudrun-golang/db/model"
)

// CounterInterface 数据模型接口
type CounterInterface interface {
	GetCounter(id int32) (*model.CounterModel, error)
	UpsertCounter(counter *model.CounterModel) error
	ClearCounter(id int32) error

	// User
	GetUserByOpenID(openid string) (*model.User, error)
	GetUserByToken(token string) (*model.User, error)
	UpsertUser(user *model.User) error

	// Lottery
	CreateLottery(lottery *model.Lottery) error
	UpdateLottery(lottery *model.Lottery) error
	DeleteLottery(id string) error
	GetActiveLotteries() ([]*model.Lottery, error)
	GetLotteryByID(id string) (*model.Lottery, error)
	GetUserCreatedLotteries(userID string) ([]*model.Lottery, error)
	GetLotteriesByAuditStatus(auditStatus string) ([]*model.Lottery, error)
	GetAllLotteries() ([]*model.Lottery, error)

	// Participant
	CreateParticipant(participant *model.LotteryParticipant) error
	UpdateParticipant(participant *model.LotteryParticipant) error
	GetUserParticipants(userID string) ([]*model.LotteryParticipant, error)
	GetLotteryParticipants(lotteryID string) ([]*model.LotteryParticipant, error)

	// Post (Legacy adapter)
	CreatePost(post *model.Post) error
	GetActivePosts() ([]*model.Post, error)
	GetUserPosts(userID string) ([]*model.Post, error)

	// Activity
	CreateActivity(activity *model.Activity) error
	UpdateActivity(activity *model.Activity) error
	DeleteActivity(id string) error
	GetActivityByID(id string) (*model.Activity, error)
	GetActiveActivities() ([]*model.Activity, error)
	GetAllActivities() ([]*model.Activity, error)
	GetUserCreatedActivities(userID string) ([]*model.Activity, error)
	GetLatestActiveActivity() (*model.Activity, error)
	GetRecentActivities(limit int) ([]*model.Activity, error)
}

// CounterInterfaceImp 数据模型实现
type CounterInterfaceImp struct{}

// Imp 实现实例
var Imp CounterInterface = &CounterInterfaceImp{}
