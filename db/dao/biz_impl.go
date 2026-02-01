package dao

import (
	"wxcloudrun-golang/db"
	"wxcloudrun-golang/db/model"
)

// GetUserByOpenID 根据OpenID获取用户
func (imp *CounterInterfaceImp) GetUserByOpenID(openid string) (*model.User, error) {
	var user model.User
	err := db.Get().Where("open_id = ?", openid).First(&user).Error
	return &user, err
}

// GetUserByToken 根据Token获取用户
func (imp *CounterInterfaceImp) GetUserByToken(token string) (*model.User, error) {
	var user model.User
	err := db.Get().Where("token = ?", token).First(&user).Error
	return &user, err
}

// UpsertUser 更新或创建用户
func (imp *CounterInterfaceImp) UpsertUser(user *model.User) error {
	return db.Get().Save(user).Error
}

// GetActiveLotteries 获取进行中的抽奖活动
func (imp *CounterInterfaceImp) GetActiveLotteries() ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	err := db.Get().Where("status = ?", 1).Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

// GetLotteryByID 获取抽奖活动详情
func (imp *CounterInterfaceImp) GetLotteryByID(id int32) (*model.Lottery, error) {
	var lottery model.Lottery
	err := db.Get().Where("id = ?", id).First(&lottery).Error
	return &lottery, err
}

// GetUserLotteries 获取用户发布的抽奖
func (imp *CounterInterfaceImp) GetUserLotteries(userID int32) ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	err := db.Get().Where("user_id = ?", userID).Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

// CreateLottery 创建抽奖活动
func (imp *CounterInterfaceImp) CreateLottery(lottery *model.Lottery) error {
	return db.Get().Create(lottery).Error
}

// CreateRecord 创建抽奖记录
func (imp *CounterInterfaceImp) CreateRecord(record *model.UserLotteryRecord) error {
	return db.Get().Create(record).Error
}

// GetUserRecords 获取用户抽奖记录
func (imp *CounterInterfaceImp) GetUserRecords(userID int32) ([]*model.UserLotteryRecord, error) {
	var records []*model.UserLotteryRecord
	err := db.Get().Where("user_id = ?", userID).Order("created_at desc").Find(&records).Error
	return records, err
}

// CreatePost 创建发布内容
func (imp *CounterInterfaceImp) CreatePost(post *model.Post) error {
	return db.Get().Create(post).Error
}

// GetActivePosts 获取已发布的帖子
func (imp *CounterInterfaceImp) GetActivePosts() ([]*model.Post, error) {
	var posts []*model.Post
	err := db.Get().Where("status = ?", 1).Order("created_at desc").Find(&posts).Error
	return posts, err
}

// GetUserPosts 获取用户发布的帖子
func (imp *CounterInterfaceImp) GetUserPosts(userID int32) ([]*model.Post, error) {
	var posts []*model.Post
	err := db.Get().Where("user_id = ?", userID).Order("created_at desc").Find(&posts).Error
	return posts, err
}

// GetPostByID 获取帖子详情
func (imp *CounterInterfaceImp) GetPostByID(id int32) (*model.Post, error) {
	var post model.Post
	err := db.Get().Where("id = ?", id).First(&post).Error
	return &post, err
}

// UpdatePost 更新帖子
func (imp *CounterInterfaceImp) UpdatePost(post *model.Post) error {
	return db.Get().Save(post).Error
}

// DeletePost 删除帖子 (软删除或硬删除，这里使用硬删除)
func (imp *CounterInterfaceImp) DeletePost(id int32) error {
	return db.Get().Delete(&model.Post{}, id).Error
}
