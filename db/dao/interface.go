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
	GetActiveLotteries() ([]*model.Lottery, error)
	GetLotteryByID(id int32) (*model.Lottery, error)
	CreateLottery(lottery *model.Lottery) error
	GetUserLotteries(userID int32) ([]*model.Lottery, error) // New

	// Record
	CreateRecord(record *model.UserLotteryRecord) error
	GetUserRecords(userID int32) ([]*model.UserLotteryRecord, error)

	// Post
	CreatePost(post *model.Post) error
	GetActivePosts() ([]*model.Post, error)
	GetUserPosts(userID int32) ([]*model.Post, error)
	GetPostByID(id int32) (*model.Post, error)
	UpdatePost(post *model.Post) error
	DeletePost(id int32) error
}

// CounterInterfaceImp 数据模型实现
type CounterInterfaceImp struct{}

// Imp 实现实例
var Imp CounterInterface = &CounterInterfaceImp{}
