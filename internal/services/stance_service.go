package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/repositories"
	"encoding/json"
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var StanceService = newStanceService()

func newStanceService() *stanceService {
	return &stanceService{}
}

type stanceService struct{}

func (s *stanceService) Get(id int64) *models.Stance {
	if id <= 0 {
		return nil
	}
	return repositories.StanceRepository.Get(sqls.DB(), id)
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

func (s *stanceService) GetStanceChoices(stanceId int64) []models.StanceChoice {
	return repositories.StanceChoiceRepository.FindByStanceId(sqls.DB(), stanceId)
}

func (s *stanceService) CreateStance(userId int64, form req.StanceCreateReq) (*models.Stance, error) {
	if form.PollId <= 0 {
		return nil, errors.New(locales.Get("vote.poll_id_required"))
	}
	if len(form.OptionIds) == 0 && !form.NoneOfTheAbove {
		return nil, errors.New(locales.Get("vote.select_option_required"))
	}
	form.Reason = strings.TrimSpace(form.Reason)

	var result *models.Stance
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx
		vote := &models.Vote{}
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(vote, "id = ?", form.PollId).Error; err != nil {
			return errors.New(locales.Get("vote.not_found"))
		}
		if vote.PollType != constants.PollTypeProposal {
			return errors.New(locales.Get("vote.proposal_only"))
		}
		if vote.ClosedAt != nil || dates.NowTimestamp() > vote.ExpiredAt {
			return errors.New(locales.Get("vote.expired"))
		}
		if vote.StanceReasonRequired == constants.StanceReasonMust && form.Reason == "" {
			return errors.New(locales.Get("vote.reason_required"))
		}
		if len(form.OptionIds) > vote.VoteNum && vote.VoteNum > 0 {
			return errors.New(locales.Get("vote.multiple_over_limit"))
		}

		options := repositories.VoteOptionRepository.Find(tx, sqls.NewCnd().Eq("vote_id", vote.Id).Asc("sort_no").Asc("id"))
		if len(options) == 0 {
			return errors.New(locales.Get("vote.option_not_found"))
		}
		optionMap := make(map[int64]models.VoteOption, len(options))
		for _, option := range options {
			optionMap[option.Id] = option
		}

		selected := make([]int64, 0, len(form.OptionIds))
		selectedSet := make(map[int64]bool, len(form.OptionIds))
		for _, optionId := range form.OptionIds {
			if optionId <= 0 {
				return errors.New(locales.Get("vote.option_invalid"))
			}
			if _, ok := optionMap[optionId]; !ok {
				return errors.New(locales.Get("vote.option_invalid"))
			}
			if selectedSet[optionId] {
				continue
			}
			selectedSet[optionId] = true
			selected = append(selected, optionId)
		}
		if len(selected) == 0 && !form.NoneOfTheAbove {
			return errors.New(locales.Get("vote.select_option_required"))
		}

		existing := repositories.StanceRepository.GetLatestByParticipant(tx, vote.Id, userId)
		if existing != nil {
			if err := tx.Model(&models.Stance{}).Where("id = ?", existing.Id).UpdateColumn("latest", false).Error; err != nil {
				return err
			}
			if err := s.updateOptionCounts(tx, existing.Id, -1); err != nil {
				return err
			}
		}

		optionScoresJSON, _ := json.Marshal(form.OptionScores)
		now := dates.NowTimestamp()
		stance := &models.Stance{
			PollId:         vote.Id,
			ParticipantId:  userId,
			Reason:         form.Reason,
			ReasonFormat:   string(constants.ContentTypeText),
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

		for _, optionId := range selected {
			score := 1
			if sc, ok := form.OptionScores[optionId]; ok && sc > 0 {
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

		if err := s.updateOptionCounts(tx, stance.Id, 1); err != nil {
			return err
		}

		voterCount := repositories.StanceRepository.CountByPollId(tx, vote.Id)
		if err := tx.Model(&models.Vote{}).Where("id = ?", vote.Id).UpdateColumn("vote_count", voterCount).Error; err != nil {
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
			return errors.New(locales.Get("vote.stance_not_found"))
		}
		if stance.ParticipantId != userId {
			return errors.New(locales.Get("vote.stance_revoke_forbidden"))
		}
		if !stance.Latest {
			return errors.New(locales.Get("vote.stance_revoke_latest_only"))
		}

		now := dates.NowTimestamp()
		if err := tx.Model(&models.Stance{}).Where("id = ?", stance.Id).Updates(map[string]interface{}{
			"latest":     false,
			"revoked_at": now,
			"revoker_id": userId,
			"update_time": now,
		}).Error; err != nil {
			return err
		}
		if err := s.updateOptionCounts(tx, stance.Id, -1); err != nil {
			return err
		}
		voterCount := repositories.StanceRepository.CountByPollId(tx, stance.PollId)
		return tx.Model(&models.Vote{}).Where("id = ?", stance.PollId).UpdateColumn("vote_count", voterCount).Error
	})
}

func (s *stanceService) CanViewResults(vote *models.Vote, currentUserId int64) bool {
	if vote == nil {
		return false
	}
	if vote.HideResults == constants.HideResultsOff {
		return true
	}
	if currentUserId <= 0 {
		return false
	}
	if vote.HideResults == constants.HideResultsUntilClosed {
		return vote.ClosedAt != nil || dates.NowTimestamp() > vote.ExpiredAt
	}
	if vote.HideResults == constants.HideResultsUntilVote {
		stance := s.GetLatestByParticipant(vote.Id, currentUserId)
		return stance != nil || vote.ClosedAt != nil || dates.NowTimestamp() > vote.ExpiredAt
	}
	return false
}

func (s *stanceService) updateOptionCounts(tx *gorm.DB, stanceId int64, delta int) error {
	choices := repositories.StanceChoiceRepository.FindByStanceId(tx, stanceId)
	for _, choice := range choices {
		if delta > 0 {
			if err := tx.Model(&models.VoteOption{}).Where("id = ?", choice.PollOptionId).Updates(map[string]interface{}{
				"vote_count":  gorm.Expr("vote_count + 1"),
				"voter_count": gorm.Expr("voter_count + 1"),
				"total_score": gorm.Expr("total_score + ?", choice.Score),
			}).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Model(&models.VoteOption{}).Where("id = ?", choice.PollOptionId).Updates(map[string]interface{}{
				"vote_count":  gorm.Expr("CASE WHEN vote_count > 0 THEN vote_count - 1 ELSE 0 END"),
				"voter_count": gorm.Expr("CASE WHEN voter_count > 0 THEN voter_count - 1 ELSE 0 END"),
				"total_score": gorm.Expr("CASE WHEN total_score >= ? THEN total_score - ? ELSE 0 END", choice.Score, choice.Score),
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}
