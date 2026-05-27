package services

import (
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"testing"
	"time"
)

func TestOutcomeService_CreateAndUpdateOutcome(t *testing.T) {
	setupProposalVoteTestDB(t)
	now := time.Now().UnixMilli()
	owner := mustCreateProposalVoteUser(t, now)
	other := mustCreateProposalVoteUser(t, now+1)
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
	options := VoteOptionService.FindByVoteId(vote.Id)
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}

	if _, err := OutcomeService.CreateOutcome(other.Id, req.OutcomeCreateReq{
		PollId:    vote.Id,
		Statement: "forbidden",
	}); err == nil {
		t.Fatal("expected non-owner create outcome to fail")
	}

	first, err := OutcomeService.CreateOutcome(owner.Id, req.OutcomeCreateReq{
		PollId:       vote.Id,
		Statement:    "first outcome",
		PollOptionId: &options[0].Id,
	})
	if err != nil {
		t.Fatalf("create outcome: %v", err)
	}
	if !first.Latest {
		t.Fatal("expected first outcome latest")
	}

	second, err := OutcomeService.CreateOutcome(owner.Id, req.OutcomeCreateReq{
		PollId:       vote.Id,
		Statement:    "second outcome",
		PollOptionId: &options[1].Id,
	})
	if err != nil {
		t.Fatalf("create second outcome: %v", err)
	}
	if !second.Latest {
		t.Fatal("expected second outcome latest")
	}
	latest := OutcomeService.GetLatestByPollId(vote.Id)
	if latest == nil || latest.Id != second.Id {
		t.Fatalf("expected latest outcome id=%d, got %+v", second.Id, latest)
	}

	updated, err := OutcomeService.UpdateOutcome(owner.Id, second.Id, req.OutcomeCreateReq{
		Statement:       "updated outcome",
		StatementFormat: "markdown",
		PollOptionId:    &options[0].Id,
	})
	if err != nil {
		t.Fatalf("update outcome: %v", err)
	}
	if updated.Statement != "updated outcome" || updated.StatementFormat != "markdown" {
		t.Fatalf("unexpected updated outcome: %+v", updated)
	}

	if _, err := OutcomeService.UpdateOutcome(other.Id, second.Id, req.OutcomeCreateReq{
		Statement: "bad",
	}); err == nil {
		t.Fatal("expected non-author update to fail")
	}
}
