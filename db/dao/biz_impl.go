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

// CreateLottery 创建抽奖活动
func (imp *CounterInterfaceImp) CreateLottery(lottery *model.Lottery) error {
	return db.Get().Create(lottery).Error
}

// UpdateLottery 更新抽奖活动
func (imp *CounterInterfaceImp) UpdateLottery(lottery *model.Lottery) error {
	return db.Get().Save(lottery).Error
}

// DeleteLottery 删除抽奖活动
func (imp *CounterInterfaceImp) DeleteLottery(id string) error {
	return db.Get().Where("id = ?", id).Delete(&model.Lottery{}).Error
}

// GetActiveLotteries 获取进行中的抽奖活动
func (imp *CounterInterfaceImp) GetActiveLotteries() ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	// Status: 'active', AuditStatus: 'approved'
	err := db.Get().Where("status = ? AND audit_status = ?", "active", "approved").Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

// GetLotteryByID 获取抽奖活动详情
func (imp *CounterInterfaceImp) GetLotteryByID(id string) (*model.Lottery, error) {
	var lottery model.Lottery
	err := db.Get().Where("id = ?", id).First(&lottery).Error
	return &lottery, err
}

// GetUserCreatedLotteries 获取用户创建的抽奖活动
func (imp *CounterInterfaceImp) GetUserCreatedLotteries(userID string) ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	err := db.Get().Where("creator_id = ?", userID).Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

func (imp *CounterInterfaceImp) GetLotteriesByAuditStatus(auditStatus string) ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	err := db.Get().Where("audit_status = ?", auditStatus).Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

func (imp *CounterInterfaceImp) GetAllLotteries() ([]*model.Lottery, error) {
	var lotteries []*model.Lottery
	err := db.Get().Order("created_at desc").Find(&lotteries).Error
	return lotteries, err
}

// CreateParticipant 创建参与记录
func (imp *CounterInterfaceImp) CreateParticipant(participant *model.LotteryParticipant) error {
	return db.Get().Create(participant).Error
}

// GetUserParticipants 获取用户参与记录
func (imp *CounterInterfaceImp) GetUserParticipants(userID string) ([]*model.LotteryParticipant, error) {
	var participants []*model.LotteryParticipant
	err := db.Get().Where("user_id = ?", userID).Order("participated_at desc").Find(&participants).Error
	return participants, err
}

// GetLotteryParticipants 获取某活动的参与者
func (imp *CounterInterfaceImp) GetLotteryParticipants(lotteryID string) ([]*model.LotteryParticipant, error) {
	var participants []*model.LotteryParticipant
	err := db.Get().Where("lottery_id = ?", lotteryID).Order("participated_at desc").Find(&participants).Error
	return participants, err
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
func (imp *CounterInterfaceImp) GetUserPosts(userID string) ([]*model.Post, error) {
	var posts []*model.Post
	err := db.Get().Where("user_id = ?", userID).Order("created_at desc").Find(&posts).Error
	return posts, err
}
