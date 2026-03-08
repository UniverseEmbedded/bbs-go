package repositories

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var StanceChoiceRepository = newStanceChoiceRepository()

func newStanceChoiceRepository() *stanceChoiceRepository {
	return &stanceChoiceRepository{}
}

type stanceChoiceRepository struct {
}

func (r *stanceChoiceRepository) Get(db *gorm.DB, id int64) *models.StanceChoice {
	ret := &models.StanceChoice{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *stanceChoiceRepository) Take(db *gorm.DB, where ...interface{}) *models.StanceChoice {
	ret := &models.StanceChoice{}
	if err := db.Take(ret, where...).Error; err != nil {
		return nil
	}
	return ret
}

func (r *stanceChoiceRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.StanceChoice) {
	cnd.Find(db, &list)
	return
}

func (r *stanceChoiceRepository) FindOne(db *gorm.DB, cnd *sqls.Cnd) *models.StanceChoice {
	ret := &models.StanceChoice{}
	if err := cnd.FindOne(db, &ret); err != nil {
		return nil
	}
	return ret
}

func (r *stanceChoiceRepository) FindPageByParams(db *gorm.DB, params *params.QueryParams) (list []models.StanceChoice, paging *sqls.Paging) {
	return r.FindPageByCnd(db, &params.Cnd)
}

func (r *stanceChoiceRepository) FindPageByCnd(db *gorm.DB, cnd *sqls.Cnd) (list []models.StanceChoice, paging *sqls.Paging) {
	cnd.Find(db, &list)
	count := cnd.Count(db, &models.StanceChoice{})

	paging = &sqls.Paging{
		Page:  cnd.Paging.Page,
		Limit: cnd.Paging.Limit,
		Total: count,
	}
	return
}

func (r *stanceChoiceRepository) Count(db *gorm.DB, cnd *sqls.Cnd) int64 {
	return cnd.Count(db, &models.StanceChoice{})
}

func (r *stanceChoiceRepository) Create(db *gorm.DB, t *models.StanceChoice) (err error) {
	err = db.Create(t).Error
	return
}

func (r *stanceChoiceRepository) Update(db *gorm.DB, t *models.StanceChoice) (err error) {
	err = db.Save(t).Error
	return
}

func (r *stanceChoiceRepository) Updates(db *gorm.DB, id int64, columns map[string]interface{}) (err error) {
	err = db.Model(&models.StanceChoice{}).Where("id = ?", id).Updates(columns).Error
	return
}

func (r *stanceChoiceRepository) UpdateColumn(db *gorm.DB, id int64, name string, value interface{}) (err error) {
	err = db.Model(&models.StanceChoice{}).Where("id = ?", id).UpdateColumn(name, value).Error
	return
}

func (r *stanceChoiceRepository) Delete(db *gorm.DB, id int64) {
	db.Delete(&models.StanceChoice{}, "id = ?", id)
}

func (r *stanceChoiceRepository) FindByStanceId(db *gorm.DB, stanceId int64) []models.StanceChoice {
	return r.Find(db, sqls.NewCnd().Eq("stance_id", stanceId))
}

func (r *stanceChoiceRepository) DeleteByStanceId(db *gorm.DB, stanceId int64) {
	db.Delete(&models.StanceChoice{}, "stance_id = ?", stanceId)
}
