package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/locales"
	"bbs-go/internal/repositories"
	"errors"
	"strings"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
)

var OutcomeService = newOutcomeService()

func newOutcomeService() *outcomeService {
	return &outcomeService{}
}

type outcomeService struct{}

func (s *outcomeService) Get(id int64) *models.Outcome {
	if id <= 0 {
		return nil
	}
	return repositories.OutcomeRepository.Get(sqls.DB(), id)
}

func (s *outcomeService) GetLatestByPollId(pollId int64) *models.Outcome {
	return repositories.OutcomeRepository.GetLatestByPollId(sqls.DB(), pollId)
}

func (s *outcomeService) CreateOutcome(authorId int64, form req.OutcomeCreateReq) (*models.Outcome, error) {
	if form.PollId <= 0 {
		return nil, errors.New(locales.Get("vote.poll_id_required"))
	}
	form.Statement = strings.TrimSpace(form.Statement)
	if form.Statement == "" {
		return nil, errors.New(locales.Get("vote.outcome_statement_required"))
	}
	if form.StatementFormat == "" {
		form.StatementFormat = "text"
	}

	vote := VoteService.Get(form.PollId)
	if vote == nil {
		return nil, errors.New(locales.Get("vote.not_found"))
	}
	if vote.UserId != authorId {
		return nil, errors.New(locales.Get("vote.outcome_forbidden"))
	}
	if form.PollOptionId != nil && VoteOptionService.Get(*form.PollOptionId) == nil {
		return nil, errors.New(locales.Get("vote.option_invalid"))
	}

	var result *models.Outcome
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx
		existing := repositories.OutcomeRepository.GetLatestByPollId(tx, form.PollId)
		if existing != nil {
			if err := tx.Model(&models.Outcome{}).Where("id = ?", existing.Id).UpdateColumn("latest", false).Error; err != nil {
				return err
			}
		}
		now := dates.NowTimestamp()
		outcome := &models.Outcome{
			PollId:          form.PollId,
			Statement:       form.Statement,
			StatementFormat: form.StatementFormat,
			AuthorId:        authorId,
			PollOptionId:    form.PollOptionId,
			Latest:          true,
			ReviewOn:        form.ReviewOn,
			CustomFields:    form.CustomFields,
			CreateTime:      now,
			UpdateTime:      now,
		}
		if err := repositories.OutcomeRepository.Create(tx, outcome); err != nil {
			return err
		}
		result = outcome
		return nil
	})
	return result, err
}

func (s *outcomeService) UpdateOutcome(authorId, outcomeId int64, form req.OutcomeCreateReq) (*models.Outcome, error) {
	outcome := s.Get(outcomeId)
	if outcome == nil {
		return nil, errors.New(locales.Get("vote.outcome_not_found"))
	}
	if outcome.AuthorId != authorId {
		return nil, errors.New(locales.Get("vote.outcome_update_forbidden"))
	}
	form.Statement = strings.TrimSpace(form.Statement)
	if form.Statement == "" {
		return nil, errors.New(locales.Get("vote.outcome_statement_required"))
	}
	if form.StatementFormat == "" {
		form.StatementFormat = outcome.StatementFormat
		if form.StatementFormat == "" {
			form.StatementFormat = "text"
		}
	}
	if form.PollOptionId != nil && VoteOptionService.Get(*form.PollOptionId) == nil {
		return nil, errors.New(locales.Get("vote.option_invalid"))
	}

	now := dates.NowTimestamp()
	outcome.Statement = form.Statement
	outcome.StatementFormat = form.StatementFormat
	outcome.PollOptionId = form.PollOptionId
	outcome.ReviewOn = form.ReviewOn
	outcome.CustomFields = form.CustomFields
	outcome.UpdateTime = now
	if err := repositories.OutcomeRepository.Update(sqls.DB(), outcome); err != nil {
		return nil, err
	}
	return outcome, nil
}
