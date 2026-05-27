package repositories

import (
	"bbs-go/internal/models"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var OutcomeRepository = newOutcomeRepository()

func newOutcomeRepository() *outcomeRepository {
	return &outcomeRepository{}
}

type outcomeRepository struct{}

func (r *outcomeRepository) Get(db *gorm.DB, id int64) *models.Outcome {
	ret := &models.Outcome{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *outcomeRepository) Take(db *gorm.DB, where ...interface{}) *models.Outcome {
	ret := &models.Outcome{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *outcomeRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.Outcome) {
	cnd.Find(db, &list)
	return
}

func (r *outcomeRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.Outcome {
	ret := &models.Outcome{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *outcomeRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.Outcome, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *outcomeRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.Outcome, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.Outcome{})
	paging = &sqls.Paging{Page: cnd.Paging.Page, Limit: cnd.Paging.Limit, Total: count}
	return
}

func (r *outcomeRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.Outcome{})
}

func (r *outcomeRepository) Create(db *gorm.DB, t *models.Outcome) error {
	return db.Create(t).Error
}

func (r *outcomeRepository) Update(db *gorm.DB, t *models.Outcome) error {
	return db.Save(t).Error
}

func (r *outcomeRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) error {
	return db.Model(&models.Outcome{}).Where("id = ?", id).Updates(columns).Error
}

func (r *outcomeRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) error {
	return db.Model(&models.Outcome{}).Where("id = ?", id).UpdateColumn(name, value).Error
}

func (r *outcomeRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.Outcome{}, "id = ?", id)
}

func (r *outcomeRepository) GetLatestByPollId(db *gorm.DB, pollId int64) *models.Outcome {
	return r.FindOne(db, sqls.NewCnd().Where("poll_id = ? AND latest = ?", pollId, true))
}
