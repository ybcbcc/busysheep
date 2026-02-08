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

func (imp *CounterInterfaceImp) GetUserByID(userID string) (*model.User, error) {
	var user model.User
	err := db.Get().Where("id = ?", userID).First(&user).Error
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

// UpdateParticipant 更新参与记录
func (imp *CounterInterfaceImp) UpdateParticipant(participant *model.LotteryParticipant) error {
	return db.Get().Save(participant).Error
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

// CreateActivity 创建活动
func (imp *CounterInterfaceImp) CreateActivity(activity *model.Activity) error {
	return db.Get().Create(activity).Error
}

// UpdateActivity 更新活动
func (imp *CounterInterfaceImp) UpdateActivity(activity *model.Activity) error {
	return db.Get().Save(activity).Error
}

// DeleteActivity 删除活动
func (imp *CounterInterfaceImp) DeleteActivity(id string) error {
	return db.Get().Where("id = ?", id).Delete(&model.Activity{}).Error
}

// GetActivityByID 获取活动详情
func (imp *CounterInterfaceImp) GetActivityByID(id string) (*model.Activity, error) {
	var activity model.Activity
	err := db.Get().Where("id = ?", id).First(&activity).Error
	return &activity, err
}

// GetActiveActivities 获取当前有效活动
func (imp *CounterInterfaceImp) GetActiveActivities() ([]*model.Activity, error) {
	var activities []*model.Activity
	// 当前时间在时间窗内且 AppearanceCount != 0
	err := db.Get().Where("start_time <= NOW() AND DATE_ADD(start_time, INTERVAL duration_minutes MINUTE) >= NOW() AND appearance_count <> 0").
		Order("created_at desc").Find(&activities).Error
	return activities, err
}

// GetAllActivities 获取全部活动
func (imp *CounterInterfaceImp) GetAllActivities() ([]*model.Activity, error) {
	var activities []*model.Activity
	err := db.Get().Order("created_at desc").Find(&activities).Error
	return activities, err
}

// GetUserCreatedActivities 获取用户创建的活动
func (imp *CounterInterfaceImp) GetUserCreatedActivities(userID string) ([]*model.Activity, error) {
	var activities []*model.Activity
	err := db.Get().Where("creator_id = ?", userID).Order("created_at desc").Find(&activities).Error
	return activities, err
}

// GetLatestActiveActivity 获取最新有效的活动广告
func (imp *CounterInterfaceImp) GetLatestActiveActivity() (*model.Activity, error) {
	var activity model.Activity
	err := db.Get().Where("start_time <= NOW() AND DATE_ADD(start_time, INTERVAL duration_minutes MINUTE) >= NOW() AND appearance_count <> 0").
		Order("updated_at desc").First(&activity).Error
	return &activity, err
}

// GetRecentActivities 获取最近活动公告
func (imp *CounterInterfaceImp) GetRecentActivities(limit int) ([]*model.Activity, error) {
	var activities []*model.Activity
	if limit <= 0 {
		limit = 10
	}
	err := db.Get().Order("updated_at desc").Limit(limit).Find(&activities).Error
	return activities, err
}

// GetActivityExposure 查询用户曝光记录
func (imp *CounterInterfaceImp) GetActivityExposure(activityID, userID string) (*model.ActivityExposure, error) {
	var exp model.ActivityExposure
	err := db.Get().Where("activity_id = ? AND user_id = ?", activityID, userID).First(&exp).Error
	return &exp, err
}

// UpsertActivityExposure 新增或更新曝光记录
func (imp *CounterInterfaceImp) UpsertActivityExposure(exp *model.ActivityExposure) error {
	return db.Get().Save(exp).Error
}
