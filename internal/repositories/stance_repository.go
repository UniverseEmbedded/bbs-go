package repositories

import (
	"bbs-go/internal/models"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var StanceRepository = newStanceRepository()

func newStanceRepository() *stanceRepository {
	return &stanceRepository{}
}

type stanceRepository struct{}

func (r *stanceRepository) Get(db *gorm.DB, id int64) *models.Stance {
	ret := &models.Stance{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *stanceRepository) Take(db *gorm.DB, where ...interface{}) *models.Stance {
	ret := &models.Stance{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *stanceRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Stance) {
	cnd.Find(db, &list)
	return
}

func (r *stanceRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Stance {
	ret := &models.Stance{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *stanceRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Stance, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *stanceRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Stance, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Stance{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *stanceRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Stance{})
}

func (r *stanceRepository) Create(db *gorm.DB, t *models.Stance) error {
	return db.Create(t).Error
}

func (r *stanceRepository) Update(db *gorm.DB, t *models.Stance) error {
	return db.Save(t).Error
}

func (r *stanceRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.Stance{}).Where("id = ?", id).Updates(columns).Error
}

func (r *stanceRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.Stance{}).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *stanceRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Stance{}, "id = ?", id)
}

func (r *stanceRepository) GetLatestByParticipant(db *gorm.DB, pollId, participantId int64) *models.Stance {
	return r.FindOne(db, sqls.NewCnd().Where("poll_id = ? AND participant_id = ? AND latest = ?", pollId, participantId, true))
}

func (r *stanceRepository) FindByPollId(db *gorm.DB, pollId int64) (list []models.Stance) {
	db.Where("poll_id = ? AND latest = ?", pollId, true).Order("create_time DESC, id DESC").Find(&list)
	return
}

func (r *stanceRepository) CountByPollId(db *gorm.DB, pollId int64) int64 {
	var count int64
	db.Model(&models.Stance{}).Where("poll_id = ? AND latest = ?", pollId, true).Count(&count)
	return count
}
