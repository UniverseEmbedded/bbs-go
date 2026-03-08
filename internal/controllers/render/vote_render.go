package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/common/dates"
)

func BuildVote(ctx iris.Context, vote *models.Vote) *resp.VoteResponse {
	if vote == nil {
		return nil
	}

	ret := &resp.VoteResponse{
		Id:                   vote.Id,
		Type:                 vote.Type,
		Title:                vote.Title,
		ExpiredAt:            vote.ExpiredAt,
		VoteNum:              vote.VoteNum,
		OptionCount:          vote.OptionCount,
		VoteCount:            vote.VoteCount,
		Expired:              vote.ClosedAt != nil || dates.NowTimestamp() > vote.ExpiredAt,
		PollType:             vote.PollType,
		HideResults:          vote.HideResults,
		Anonymous:            vote.Anonymous,
		OpeningAt:            vote.OpeningAt,
		OpenedAt:             vote.OpenedAt,
		ClosingAt:            vote.ClosingAt,
		ClosedAt:             vote.ClosedAt,
		VoterCanAddOptions:   vote.VoterCanAddOptions,
		SpecifiedVotersOnly:  vote.SpecifiedVotersOnly,
		StanceReasonRequired: vote.StanceReasonRequired,
		QuorumPct:            vote.QuorumPct,
	}

	currentUserId := common.GetCurrentUserID(ctx)

	var selectedMap map[int64]bool
	if currentUserId > 0 {
		stance := services.StanceService.GetLatestByParticipant(vote.Id, currentUserId)
		if stance != nil {
			ret.Voted = true
			choices := services.StanceService.GetStanceChoices(stance.Id)
			ret.OptionIds = make([]int64, 0, len(choices))
			selectedMap = make(map[int64]bool, len(choices))
			for _, choice := range choices {
				ret.OptionIds = append(ret.OptionIds, choice.PollOptionId)
				selectedMap[choice.PollOptionId] = true
			}
		}
	}

	ret.CanViewResults = canViewResults(ctx, vote, currentUserId)

	if ret.CanViewResults {
		options := services.VoteOptionService.FindByVoteId(vote.Id)
		for _, option := range options {
			item := resp.VoteOptionResponse{
				Id:         option.Id,
				Content:    option.Content,
				SortNo:     option.SortNo,
				VoteCount:  option.VoteCount,
				Icon:       option.Icon,
				Meaning:    option.Meaning,
				Prompt:     option.Prompt,
				Priority:   option.Priority,
				TotalScore: option.TotalScore,
				VoterCount: option.VoterCount,
			}
			if vote.VoteCount > 0 {
				item.Percent = float64(option.VoteCount) / float64(vote.VoteCount) * 100
			}
			if selectedMap != nil {
				item.Voted = selectedMap[option.Id]
			}
			ret.Options = append(ret.Options, item)
		}
	}

	outcome := services.OutcomeService.GetLatestByPollId(vote.Id)
	if outcome != nil {
		ret.Outcome = BuildOutcome(ctx, outcome)
	}

	return ret
}

func canViewResults(ctx iris.Context, vote *models.Vote, currentUserId int64) bool {
	if vote.HideResults == constants.HideResultsOff {
		return true
	}

	if currentUserId <= 0 {
		return false
	}

	if vote.HideResults == constants.HideResultsUntilClosed {
		return vote.ClosedAt != nil
	}

	if vote.HideResults == constants.HideResultsUntilVote {
		stance := services.StanceService.GetLatestByParticipant(vote.Id, currentUserId)
		return stance != nil || vote.ClosedAt != nil
	}

	return false
}
