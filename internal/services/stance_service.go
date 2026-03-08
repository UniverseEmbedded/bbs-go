package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/repositories"
	"encoding/json"
	"errors"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var StanceService = newStanceService()

func newStanceService() *stanceService {
	return &stanceService{}
}

type stanceService struct {
}

func (s *stanceService) Get(id int64) *models.Stance {
	if id <= 0 {
		return nil
	}
	return repositories.StanceRepository.Get(sqls.DB(), id)
}

func (s *stanceService) Take(where ...interface{}) *models.Stance {
	return repositories.StanceRepository.Take(sqls.DB(), where...)
}

func (s *stanceService) Find(cnd *sqls.Cnd) []models.Stance {
	return repositories.StanceRepository.Find(sqls.DB(), cnd)
}

func (s *stanceService) FindOne(cnd *sqls.Cnd) *models.Stance {
	return repositories.StanceRepository.FindOne(sqls.DB(), cnd)
}

func (s *stanceService) FindPageByParams(params *params.QueryParams) (list []models.Stance, paging *sqls.Paging) {
	return repositories.StanceRepository.FindPageByParams(sqls.DB(), params)
}

func (s *stanceService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Stance, paging *sqls.Paging) {
	return repositories.StanceRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *stanceService) Count(cnd *sqls.Cnd) int64 {
	return repositories.StanceRepository.Count(sqls.DB(), cnd)
}

func (s *stanceService) Create(t *models.Stance) error {
	return repositories.StanceRepository.Create(sqls.DB(), t)
}

func (s *stanceService) Update(t *models.Stance) error {
	return repositories.StanceRepository.Update(sqls.DB(), t)
}

func (s *stanceService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.StanceRepository.Updates(sqls.DB(), id, columns)
}

func (s *stanceService) Delete(id int64) {
	repositories.StanceRepository.Delete(sqls.DB(), id)
}

func (s *stanceService) GetLatestByParticipant(pollId, participantId int64) *models.Stance {
	return repositories.StanceRepository.GetLatestByParticipant(sqls.DB(), pollId, participantId)
}

func (s *stanceService) FindByPollId(pollId int64) []models.Stance {
	return repositories.StanceRepository.FindByPollId(sqls.DB(), pollId)
}

func (s *stanceService) CountByPollId(pollId int64) int64 {
	return repositories.StanceRepository.CountByPollId(sqls.DB(), pollId)
}

type StanceCreateForm struct {
	PollId         int64         `json:"pollId"`
	OptionIds      []int64       `json:"optionIds"`
	OptionScores   map[int64]int `json:"optionScores"`
	Reason         string        `json:"reason"`
	NoneOfTheAbove bool          `json:"noneOfTheAbove"`
}

func (s *stanceService) CreateStance(userId int64, form StanceCreateForm) (*models.Stance, error) {
	if form.PollId <= 0 {
		return nil, errors.New("pollId不能为空")
	}

	var result *models.Stance
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx

		vote := &models.Vote{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(vote, "id = ?", form.PollId).Error; err != nil {
			return errors.New("投票不存在")
		}

		if vote.ClosedAt != nil {
			return errors.New("投票已关闭")
		}

		if vote.StanceReasonRequired == constants.StanceReasonMust && form.Reason == "" {
			return errors.New("请填写投票理由")
		}

		options := repositories.VoteOptionRepository.Find(tx, sqls.NewCnd().Eq("vote_id", vote.Id))
		if len(options) == 0 {
			return errors.New("投票选项不存在")
		}

		optionMap := make(map[int64]*models.VoteOption, len(options))
		for i := range options {
			optionMap[options[i].Id] = &options[i]
		}

		for _, optionId := range form.OptionIds {
			if _, ok := optionMap[optionId]; !ok {
				return errors.New("投票选项不合法")
			}
		}

		now := dates.NowTimestamp()

		existingStance := repositories.StanceRepository.GetLatestByParticipant(tx, vote.Id, userId)
		if existingStance != nil {
			if err := tx.Model(&models.Stance{}).Where("id = ?", existingStance.Id).
				UpdateColumn("latest", false).Error; err != nil {
				return err
			}

			if err := s.updateOptionCounts(tx, vote.Id, existingStance.Id, -1); err != nil {
				return err
			}
		}

		optionScoresJSON, _ := json.Marshal(form.OptionScores)

		stance := &models.Stance{
			PollId:         vote.Id,
			ParticipantId:  userId,
			Reason:         form.Reason,
			ReasonFormat:   "text",
			Latest:         true,
			CastAt:         &now,
			OptionScores:   string(optionScoresJSON),
			NoneOfTheAbove: form.NoneOfTheAbove,
			CreateTime:     now,
			UpdateTime:     now,
		}
		if err := repositories.StanceRepository.Create(tx, stance); err != nil {
			return err
		}

		for _, optionId := range form.OptionIds {
			score := 1
			if sc, ok := form.OptionScores[optionId]; ok {
				score = sc
			}
			choice := &models.StanceChoice{
				StanceId:     stance.Id,
				PollOptionId: optionId,
				Score:        score,
				CreateTime:   now,
			}
			if err := repositories.StanceChoiceRepository.Create(tx, choice); err != nil {
				return err
			}
		}

		if err := s.updateOptionCounts(tx, vote.Id, stance.Id, 1); err != nil {
			return err
		}

		voterCount := repositories.StanceRepository.CountByPollId(tx, vote.Id)
		if err := tx.Model(&models.Vote{}).Where("id = ?", vote.Id).
			UpdateColumn("vote_count", voterCount).Error; err != nil {
			return err
		}

		result = stance
		return nil
	})
	return result, err
}

func (s *stanceService) RevokeStance(userId, stanceId int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx

		stance := repositories.StanceRepository.Get(tx, stanceId)
		if stance == nil {
			return errors.New("立场不存在")
		}

		if stance.ParticipantId != userId {
			return errors.New("无权撤销此立场")
		}

		if !stance.Latest {
			return errors.New("只能撤销最新立场")
		}

		now := dates.NowTimestamp()
		if err := tx.Model(&models.Stance{}).Where("id = ?", stanceId).
			Updates(map[string]interface{}{
				"latest":     false,
				"revoked_at": now,
				"revoker_id": userId,
			}).Error; err != nil {
			return err
		}

		if err := s.updateOptionCounts(tx, stance.PollId, stanceId, -1); err != nil {
			return err
		}

		voterCount := repositories.StanceRepository.CountByPollId(tx, stance.PollId)
		return tx.Model(&models.Vote{}).Where("id = ?", stance.PollId).
			UpdateColumn("vote_count", voterCount).Error
	})
}

func (s *stanceService) updateOptionCounts(tx *gorm.DB, pollId, stanceId int64, delta int) error {
	choices := repositories.StanceChoiceRepository.FindByStanceId(tx, stanceId)
	for _, choice := range choices {
		if delta > 0 {
			if err := tx.Model(&models.VoteOption{}).Where("id = ?", choice.PollOptionId).
				Updates(map[string]interface{}{
					"vote_count":  gorm.Expr("vote_count + 1"),
					"voter_count": gorm.Expr("voter_count + 1"),
					"total_score": gorm.Expr("total_score + ?", choice.Score),
				}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.VoteOption{}).Where("id = ? AND vote_count > 0", choice.PollOptionId).
				Updates(map[string]interface{}{
					"vote_count":  gorm.Expr("vote_count - 1"),
					"voter_count": gorm.Expr("voter_count - 1"),
					"total_score": gorm.Expr("total_score - ?", choice.Score),
				}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *stanceService) GetStanceChoices(stanceId int64) []models.StanceChoice {
	return repositories.StanceChoiceRepository.FindByStanceId(sqls.DB(), stanceId)
}

func (s *stanceService) CanViewResults(ctx iris.Context, vote *models.Vote) bool {
	if vote.HideResults == constants.HideResultsOff {
		return true
	}

	currentUserId := common.GetCurrentUserID(ctx)
	if currentUserId <= 0 {
		return false
	}

	if vote.HideResults == constants.HideResultsUntilClosed {
		return vote.ClosedAt != nil
	}

	if vote.HideResults == constants.HideResultsUntilVote {
		stance := s.GetLatestByParticipant(vote.Id, currentUserId)
		return stance != nil || vote.ClosedAt != nil
	}

	return false
}
