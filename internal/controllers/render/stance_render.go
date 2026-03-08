package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"

	"github.com/kataras/iris/v12"
)

func BuildStance(ctx iris.Context, stance *models.Stance) *resp.StanceResponse {
	if stance == nil {
		return nil
	}

	currentUserId := common.GetCurrentUserID(ctx)

	ret := &resp.StanceResponse{
		Id:             stance.Id,
		PollId:         stance.PollId,
		ParticipantId:  stance.ParticipantId,
		Reason:         stance.Reason,
		ReasonFormat:   stance.ReasonFormat,
		Latest:         stance.Latest,
		CastAt:         stance.CastAt,
		RevokedAt:      stance.RevokedAt,
		NoneOfTheAbove: stance.NoneOfTheAbove,
		CreateTime:     stance.CreateTime,
	}

	vote := services.VoteService.Get(stance.PollId)
	if vote != nil && !vote.Anonymous {
		ret.Participant = BuildUserInfoDefaultIfNull(stance.ParticipantId)
	}

	if currentUserId > 0 && stance.ParticipantId == currentUserId {
		choices := services.StanceService.GetStanceChoices(stance.Id)
		ret.Choices = make([]resp.StanceChoiceResponse, 0, len(choices))
		for _, choice := range choices {
			ret.Choices = append(ret.Choices, resp.StanceChoiceResponse{
				Id:           choice.Id,
				StanceId:     choice.StanceId,
				PollOptionId: choice.PollOptionId,
				Score:        choice.Score,
			})
		}
	}

	return ret
}

func BuildStanceList(ctx iris.Context, stances []models.Stance) []resp.StanceResponse {
	if len(stances) == 0 {
		return nil
	}
	ret := make([]resp.StanceResponse, 0, len(stances))
	for _, stance := range stances {
		ret = append(ret, *BuildStance(ctx, &stance))
	}
	return ret
}

func BuildOutcome(ctx iris.Context, outcome *models.Outcome) *resp.OutcomeResponse {
	if outcome == nil {
		return nil
	}

	ret := &resp.OutcomeResponse{
		Id:              outcome.Id,
		PollId:          outcome.PollId,
		Statement:       outcome.Statement,
		StatementFormat: outcome.StatementFormat,
		AuthorId:        outcome.AuthorId,
		PollOptionId:    outcome.PollOptionId,
		Latest:          outcome.Latest,
		ReviewOn:        outcome.ReviewOn,
		CreateTime:      outcome.CreateTime,
		UpdateTime:      outcome.UpdateTime,
	}

	ret.Author = BuildUserInfoDefaultIfNull(outcome.AuthorId)

	if outcome.PollOptionId != nil {
		option := services.VoteOptionService.Get(*outcome.PollOptionId)
		if option != nil {
			ret.PollOption = &resp.VoteOptionResponse{
				Id:        option.Id,
				Content:   option.Content,
				SortNo:    option.SortNo,
				VoteCount: option.VoteCount,
				Icon:      option.Icon,
				Meaning:   option.Meaning,
			}
		}
	}

	return ret
}
