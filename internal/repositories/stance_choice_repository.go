package repositories

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

var StanceChoiceRepository = newStanceChoiceRepository()

func newStanceChoiceRepository() *stanceChoiceRepository {
	return &stanceChoiceRepository{}
}

type stanceChoiceRepository struct{}

func (r *stanceChoiceRepository) Get(db *gorm.DB, id int64) *models.StanceChoice {
	ret := &models.StanceChoice{}
	if err := db.First(ret, "id = ?", id).Error; err != nil {
		return nil
	}
	return ret
}

func (r *stanceChoiceRepository) Find(db *gorm.DB, cnd *sqls.Cnd) (list []models.StanceChoice) {
	cnd.Find(db, &list)
	return
}

func (r *stanceChoiceRepository) Create(db *gorm.DB, t *models.StanceChoice) error {
	return db.Create(t).Error
}

func (r *stanceChoiceRepository) FindByStanceId(db *gorm.DB, stanceId int64) (list []models.StanceChoice) {
	db.Where("stance_id = ?", stanceId).Order("id ASC").Find(&list)
	return
}

func (r *stanceChoiceRepository) DeleteByStanceId(db *gorm.DB, stanceId int64) error {
	return db.Where("stance_id = ?", stanceId).Delete(&models.StanceChoice{}).Error
}
