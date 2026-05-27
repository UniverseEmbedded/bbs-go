package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/repositories"
	"testing"
	"time"

	"github.com/mlogclub/simple/sqls"
)

func TestStanceService_CreateStance_ReplacesLatestAndUpdatesCounts(t *testing.T) {
	setupProposalVoteTestDB(t)
	now := time.Now().UnixMilli()
	owner := mustCreateProposalVoteUser(t, now)
	topic := mustCreateProposalVoteTopic(t, owner.Id, now)
	vote := mustCreateProposalVote(t, topic.Id, owner.Id, &req.VoteDTO{
		PollType:             constants.PollTypeProposal,
		Type:                 constants.VoteTypeMultiple,
		Title:                "proposal",
		ExpiredAt:            now + int64(time.Hour/time.Millisecond),
		VoteNum:              2,
		HideResults:          constants.HideResultsUntilVote,
		StanceReasonRequired: constants.StanceReasonOptional,
		Options: []req.VoteOptionDTO{{Content: "A"}, {Content: "B"}},
	}, now)
	voter := mustCreateProposalVoteUser(t, now+1)
	options := VoteOptionService.FindByVoteId(vote.Id)
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}

	first, err := StanceService.CreateStance(voter.Id, req.StanceCreateReq{
		PollId:       vote.Id,
		OptionIds:    []int64{options[0].Id},
		OptionScores: map[int64]int{options[0].Id: 2},
		Reason:       "first",
	})
	if err != nil {
		t.Fatalf("create first stance: %v", err)
	}
	if !first.Latest {
		t.Fatal("expected first stance latest")
	}

	voteAfterFirst := VoteService.Get(vote.Id)
	if voteAfterFirst.VoteCount != 1 {
		t.Fatalf("expected vote_count=1 after first stance, got %d", voteAfterFirst.VoteCount)
	}
	optionA := VoteOptionService.Get(options[0].Id)
	if optionA.VoterCount != 1 || optionA.TotalScore != 2 {
		t.Fatalf("expected option A voterCount=1 totalScore=2, got voterCount=%d totalScore=%d", optionA.VoterCount, optionA.TotalScore)
	}

	second, err := StanceService.CreateStance(voter.Id, req.StanceCreateReq{
		PollId:       vote.Id,
		OptionIds:    []int64{options[1].Id},
		OptionScores: map[int64]int{options[1].Id: 3},
		Reason:       "second",
	})
	if err != nil {
		t.Fatalf("create second stance: %v", err)
	}
	if !second.Latest {
		t.Fatal("expected second stance latest")
	}

	old := repositories.StanceRepository.Get(sqls.DB(), first.Id)
	if old == nil || old.Latest {
		t.Fatal("expected first stance to be marked non-latest")
	}
	optionA = VoteOptionService.Get(options[0].Id)
	optionB := VoteOptionService.Get(options[1].Id)
	if optionA.VoterCount != 0 || optionA.TotalScore != 0 {
		t.Fatalf("expected option A reset to zero, got voterCount=%d totalScore=%d", optionA.VoterCount, optionA.TotalScore)
	}
	if optionB.VoterCount != 1 || optionB.TotalScore != 3 {
		t.Fatalf("expected option B voterCount=1 totalScore=3, got voterCount=%d totalScore=%d", optionB.VoterCount, optionB.TotalScore)
	}
	voteAfterSecond := VoteService.Get(vote.Id)
	if voteAfterSecond.VoteCount != 1 {
		t.Fatalf("expected vote_count remain 1 after replacement, got %d", voteAfterSecond.VoteCount)
	}
}

func TestStanceService_RevokeStance_UpdatesLatestAndCounts(t *testing.T) {
	setupProposalVoteTestDB(t)
	now := time.Now().UnixMilli()
	owner := mustCreateProposalVoteUser(t, now)
	topic := mustCreateProposalVoteTopic(t, owner.Id, now)
	vote := mustCreateProposalVote(t, topic.Id, owner.Id, &req.VoteDTO{
		PollType:             constants.PollTypeProposal,
		Type:                 constants.VoteTypeMultiple,
		Title:                "proposal",
		ExpiredAt:            now + int64(time.Hour/time.Millisecond),
		VoteNum:              2,
		HideResults:          constants.HideResultsUntilVote,
		StanceReasonRequired: constants.StanceReasonOptional,
		Options: []req.VoteOptionDTO{{Content: "A"}, {Content: "B"}},
	}, now)
	voter := mustCreateProposalVoteUser(t, now+1)
	options := VoteOptionService.FindByVoteId(vote.Id)

	stance, err := StanceService.CreateStance(voter.Id, req.StanceCreateReq{
		PollId:    vote.Id,
		OptionIds: []int64{options[0].Id, options[1].Id},
		Reason:    "stance",
	})
	if err != nil {
		t.Fatalf("create stance: %v", err)
	}

	if err := StanceService.RevokeStance(voter.Id, stance.Id); err != nil {
		t.Fatalf("revoke stance: %v", err)
	}

	updated := repositories.StanceRepository.Get(sqls.DB(), stance.Id)
	if updated == nil || updated.Latest {
		t.Fatal("expected revoked stance to be non-latest")
	}
	if updated.RevokedAt == nil {
		t.Fatal("expected revoked_at to be set")
	}
	voteAfterRevoke := VoteService.Get(vote.Id)
	if voteAfterRevoke.VoteCount != 0 {
		t.Fatalf("expected vote_count=0 after revoke, got %d", voteAfterRevoke.VoteCount)
	}
	for _, option := range options {
		got := VoteOptionService.Get(option.Id)
		if got.VoterCount != 0 || got.TotalScore != 0 || got.VoteCount != 0 {
			t.Fatalf("expected zero counts after revoke for option %d, got voteCount=%d voterCount=%d totalScore=%d", option.Id, got.VoteCount, got.VoterCount, got.TotalScore)
		}
	}
}
