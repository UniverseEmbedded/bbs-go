package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/repositories"
	"errors"

	"github.com/mlogclub/simple/common/dates"
	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
)

var OutcomeService = newOutcomeService()

func newOutcomeService() *outcomeService {
	return &outcomeService{}
}

type outcomeService struct {
}

func (s *outcomeService) Get(id int64) *models.Outcome {
	if id <= 0 {
		return nil
	}
	return repositories.OutcomeRepository.Get(sqls.DB(), id)
}

func (s *outcomeService) Take(where ...interface{}) *models.Outcome {
	return repositories.OutcomeRepository.Take(sqls.DB(), where...)
}

func (s *outcomeService) Find(cnd *sqls.Cnd) []models.Outcome {
	return repositories.OutcomeRepository.Find(sqls.DB(), cnd)
}

func (s *outcomeService) FindOne(cnd *sqls.Cnd) *models.Outcome {
	return repositories.OutcomeRepository.FindOne(sqls.DB(), cnd)
}

func (s *outcomeService) FindPageByParams(params *params.QueryParams) (list []models.Outcome, paging *sqls.Paging) {
	return repositories.OutcomeRepository.FindPageByParams(sqls.DB(), params)
}

func (s *outcomeService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Outcome, paging *sqls.Paging) {
	return repositories.OutcomeRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *outcomeService) Count(cnd *sqls.Cnd) int64 {
	return repositories.OutcomeRepository.Count(sqls.DB(), cnd)
}

func (s *outcomeService) Create(t *models.Outcome) error {
	return repositories.OutcomeRepository.Create(sqls.DB(), t)
}

func (s *outcomeService) Update(t *models.Outcome) error {
	return repositories.OutcomeRepository.Update(sqls.DB(), t)
}

func (s *outcomeService) Updates(id int64, columns map[string]interface{}) error {
	return repositories.OutcomeRepository.Updates(sqls.DB(), id, columns)
}

func (s *outcomeService) Delete(id int64) {
	repositories.OutcomeRepository.Delete(sqls.DB(), id)
}

func (s *outcomeService) GetLatestByPollId(pollId int64) *models.Outcome {
	return repositories.OutcomeRepository.GetLatestByPollId(sqls.DB(), pollId)
}

type OutcomeCreateForm struct {
	PollId          int64  `json:"pollId"`
	Statement       string `json:"statement"`
	StatementFormat string `json:"statementFormat"`
	PollOptionId    *int64 `json:"pollOptionId"`
	ReviewOn        *int64 `json:"reviewOn"`
	CustomFields    string `json:"customFields"`
}

func (s *outcomeService) CreateOutcome(authorId int64, form OutcomeCreateForm) (*models.Outcome, error) {
	if form.PollId <= 0 {
		return nil, errors.New("pollId不能为空")
	}

	vote := VoteService.Get(form.PollId)
	if vote == nil {
		return nil, errors.New("投票不存在")
	}

	if vote.UserId != authorId {
		return nil, errors.New("只有投票创建者可以发布结果声明")
	}

	var result *models.Outcome
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx

		existingOutcome := repositories.OutcomeRepository.GetLatestByPollId(tx, form.PollId)
		if existingOutcome != nil {
			if err := tx.Model(&models.Outcome{}).Where("id = ?", existingOutcome.Id).
				UpdateColumn("latest", false).Error; err != nil {
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

func (s *outcomeService) UpdateOutcome(authorId, outcomeId int64, form OutcomeCreateForm) (*models.Outcome, error) {
	outcome := s.Get(outcomeId)
	if outcome == nil {
		return nil, errors.New("结果声明不存在")
	}

	if outcome.AuthorId != authorId {
		return nil, errors.New("只有声明创建者可以修改")
	}

	now := dates.NowTimestamp()
	outcome.Statement = form.Statement
	outcome.StatementFormat = form.StatementFormat
	outcome.PollOptionId = form.PollOptionId
	outcome.ReviewOn = form.ReviewOn
	outcome.CustomFields = form.CustomFields
	outcome.UpdateTime = now

	if err := s.Update(outcome); err != nil {
		return nil, err
	}

	return outcome, nil
}
