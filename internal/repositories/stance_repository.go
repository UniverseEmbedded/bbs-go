package repositories

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var StanceRepository = newStanceRepository()

func newStanceRepository() *stanceRepository {
	return &stanceRepository{}
}

type stanceRepository struct {
}

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

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *stanceRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Stance{})
}

func (r *stanceRepository) Create(db *gorm.DB, t *models.Stance) (err error) {
	err = db.Create(t).Error
	return
}

func (r *stanceRepository) Update(db *gorm.DB, t *models.Stance) (err error) {
	err = db.Save(t).Error
	return
}

func (r *stanceRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.Stance{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *stanceRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.Stance{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *stanceRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Stance{}, "id = ?", id)
}

func (r *stanceRepository) GetLatestByParticipant(db *gorm.DB, pollId, participantId int64) *models.Stance {
	return r.FindOne(db, sqls.NewCnd().Eq("poll_id", pollId).Eq("participant_id", participantId).Eq("latest", true).Desc("id"))
}

func (r *stanceRepository) FindByPollId(db *gorm.DB, pollId int64) []models.Stance {
	return r.Find(db, sqls.NewCnd().Eq("poll_id", pollId).Eq("latest", true).Desc("create_time"))
}

func (r *stanceRepository) CountByPollId(db *gorm.DB, pollId int64) int64 {
	return r.Count(db, sqls.NewCnd().Eq("poll_id", pollId).Eq("latest", true))
}
