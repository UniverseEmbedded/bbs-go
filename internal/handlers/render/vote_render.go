package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/common/dates"
)

func BuildVote(ctx *gin.Context, vote *models.Vote) *resp.VoteResponse {
	if vote == nil {
		return nil
	}

	ret := &resp.VoteResponse{
		Id:                   vote.Id,
		Type:                 vote.Type,
		PollType:             vote.PollType,
		Title:                vote.Title,
		ExpiredAt:            vote.ExpiredAt,
		VoteNum:              vote.VoteNum,
		OptionCount:          vote.OptionCount,
		VoteCount:            vote.VoteCount,
		Expired:              vote.ClosedAt != nil || dates.NowTimestamp() > vote.ExpiredAt,
		Anonymous:            vote.Anonymous,
		HideResults:          vote.HideResults,
		ClosedAt:             vote.ClosedAt,
		StanceReasonRequired: vote.StanceReasonRequired,
	}
	if ret.PollType == "" {
		ret.PollType = constants.PollTypePoll
	}

	currentUserId := common.GetCurrentUserID(ctx)
	selectedMap := make(map[int64]bool)

	if ret.PollType == constants.PollTypeProposal {
		if currentUserId > 0 {
			if stance := services.StanceService.GetLatestByParticipant(vote.Id, currentUserId); stance != nil {
				ret.Voted = true
				choices := services.StanceService.GetStanceChoices(stance.Id)
				ret.OptionIds = make([]int64, 0, len(choices))
				for _, choice := range choices {
					ret.OptionIds = append(ret.OptionIds, choice.PollOptionId)
					selectedMap[choice.PollOptionId] = true
				}
			}
		}
		ret.CanViewResults = services.StanceService.CanViewResults(vote, currentUserId)
		if ret.CanViewResults {
			options := services.VoteOptionService.FindByVoteId(vote.Id)
			for _, option := range options {
				item := resp.VoteOptionResponse{
					Id:         option.Id,
					Content:    option.Content,
					SortNo:     option.SortNo,
					VoteCount:  option.VoteCount,
					VoterCount: option.VoterCount,
					TotalScore: option.TotalScore,
				}
				if option.Meaning != nil {
					item.Meaning = *option.Meaning
				}
				if option.Prompt != nil {
					item.Prompt = *option.Prompt
				}
				if vote.VoteCount > 0 {
					item.Percent = float64(option.VoteCount) / float64(vote.VoteCount) * 100
				}
				if selectedMap[option.Id] {
					item.Voted = true
				}
				ret.Options = append(ret.Options, item)
			}
		}
		if outcome := services.OutcomeService.GetLatestByPollId(vote.Id); outcome != nil {
			ret.Outcome = BuildOutcome(ctx, outcome)
		}
		return ret
	}

	if currentUserId > 0 {
		if record := services.VoteRecordService.GetBy(currentUserId, vote.Id); record != nil {
			ret.Voted = true
			ret.OptionIds = services.VoteService.ParseOptionIds(record.OptionIds)
			for _, optionId := range ret.OptionIds {
				selectedMap[optionId] = true
			}
		}
	}
	ret.CanViewResults = true
	options := services.VoteOptionService.FindByVoteId(vote.Id)
	for _, option := range options {
		item := resp.VoteOptionResponse{
			Id:        option.Id,
			Content:   option.Content,
			SortNo:    option.SortNo,
			VoteCount: option.VoteCount,
		}
		if vote.VoteCount > 0 {
			item.Percent = float64(option.VoteCount) / float64(vote.VoteCount) * 100
		}
		if selectedMap[option.Id] {
			item.Voted = true
		}
		ret.Options = append(ret.Options, item)
	}
	return ret
}

func BuildStance(ctx *gin.Context, stance *models.Stance) *resp.StanceResponse {
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

func BuildStanceList(ctx *gin.Context, stances []models.Stance) []resp.StanceResponse {
	if len(stances) == 0 {
		return nil
	}
	ret := make([]resp.StanceResponse, 0, len(stances))
	for _, stance := range stances {
		if built := BuildStance(ctx, &stance); built != nil {
			ret = append(ret, *built)
		}
	}
	return ret
}

func BuildOutcome(ctx *gin.Context, outcome *models.Outcome) *resp.OutcomeResponse {
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
		Author:          BuildUserInfoDefaultIfNull(outcome.AuthorId),
	}
	if outcome.PollOptionId != nil {
		option := services.VoteOptionService.Get(*outcome.PollOptionId)
		if option != nil {
			ret.PollOption = &resp.VoteOptionResponse{
				Id:         option.Id,
				Content:    option.Content,
				SortNo:     option.SortNo,
				VoteCount:  option.VoteCount,
				VoterCount: option.VoterCount,
				TotalScore: option.TotalScore,
			}
			if option.Meaning != nil {
				ret.PollOption.Meaning = *option.Meaning
			}
			if option.Prompt != nil {
				ret.PollOption.Prompt = *option.Prompt
			}
		}
	}
	return ret
}
