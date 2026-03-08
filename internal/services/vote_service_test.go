package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/repositories"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func setupVoteTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:vote_test_%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
	sqls.SetDB(db)
	if err := db.AutoMigrate(
		&models.User{},
		&models.Vote{},
		&models.VoteOption{},
		&models.VoteRecord{},
		&models.Stance{},
		&models.StanceChoice{},
		&models.Outcome{},
	); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func mustCreateVoteUser(t *testing.T, now int64) *models.User {
	t.Helper()
	u := &models.User{
		Nickname:   "u",
		Status:     constants.StatusOk,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := repositories.UserRepository.Create(sqls.DB(), u); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return u
}

func mustCreateVote(t *testing.T, userId int64, expiredAt int64, now int64) *models.Vote {
	t.Helper()
	vote := &models.Vote{
		Type:                 constants.VoteTypeSingle,
		Title:                "Test Vote",
		ExpiredAt:            expiredAt,
		UserId:               userId,
		VoteNum:              1,
		OptionCount:          2,
		VoteCount:            0,
		CreateTime:           now,
		PollType:             constants.PollTypePoll,
		HideResults:          constants.HideResultsOff,
		Anonymous:            false,
		StanceReasonRequired: constants.StanceReasonOptional,
	}
	if err := repositories.VoteRepository.Create(sqls.DB(), vote); err != nil {
		t.Fatalf("create vote: %v", err)
	}
	return vote
}

func mustCreateVoteOption(t *testing.T, voteId int64, content string, now int64) *models.VoteOption {
	t.Helper()
	option := &models.VoteOption{
		VoteId:     voteId,
		Content:    content,
		SortNo:     1,
		CreateTime: now,
	}
	if err := repositories.VoteOptionRepository.Create(sqls.DB(), option); err != nil {
		t.Fatalf("create vote option: %v", err)
	}
	return option
}

func TestStanceService_CreateStance(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user.Id, now+86400, now)
	opt1 := mustCreateVoteOption(t, vote.Id, "Option 1", now)
	opt2 := mustCreateVoteOption(t, vote.Id, "Option 2", now)

	form := StanceCreateForm{
		PollId:    vote.Id,
		OptionIds: []int64{opt1.Id, opt2.Id},
		Reason:    "test reason",
	}
	stance, err := StanceService.CreateStance(user.Id, form)
	if err != nil {
		t.Fatalf("create stance: %v", err)
	}
	if stance == nil {
		t.Fatalf("expected stance to be created")
	}
	if stance.PollId != vote.Id {
		t.Errorf("expected pollId %d, got %d", vote.Id, stance.PollId)
	}
	if stance.ParticipantId != user.Id {
		t.Errorf("expected participantId %d, got %d", user.Id, stance.ParticipantId)
	}
	if stance.Reason != "test reason" {
		t.Errorf("expected reason 'test reason', got '%s'", stance.Reason)
	}
	if !stance.Latest {
		t.Errorf("expected latest true, got %v", stance.Latest)
	}
	if stance.CastAt == nil {
		t.Errorf("expected castAt to be set")
	}

	choices := StanceService.GetStanceChoices(stance.Id)
	if len(choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(choices))
	}
	if choices[0].PollOptionId != opt1.Id {
		t.Errorf("expected first choice to be for option 1")
	}
	if choices[0].Score != 1 {
		t.Errorf("expected score 1 for choice 1")
	}
	if choices[1].PollOptionId != opt2.Id {
		t.Errorf("expected second choice to be for option 2")
	}
	if choices[1].Score != 1 {
		t.Errorf("expected score 1 for choice 2")
	}

	voteCheck := VoteService.Get(vote.Id)
	if voteCheck == nil {
		t.Fatalf("vote should still exist")
	}
	if voteCheck.VoteCount != 1 {
		t.Errorf("expected vote count 1, got %d", voteCheck.VoteCount)
	}

	opt1Check := VoteOptionService.Get(opt1.Id)
	if opt1Check == nil {
		t.Fatalf("option 1 should still exist")
	}
	if opt1Check.VoteCount != 1 {
		t.Errorf("expected option 1 vote count 1, got %d", opt1Check.VoteCount)
	}

	opt2Check := VoteOptionService.Get(opt2.Id)
	if opt2Check == nil {
		t.Fatalf("option 2 should still exist")
	}
	if opt2Check.VoteCount != 1 {
		t.Errorf("expected option 2 vote count 1, got %d", opt2Check.VoteCount)
	}
}

func TestStanceService_CreateStance_WithReasonRequired(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)

	vote := &models.Vote{
		Type:                 constants.VoteTypeSingle,
		Title:                "Test Vote",
		ExpiredAt:            now + 86400,
		UserId:               user.Id,
		VoteNum:              1,
		OptionCount:          2,
		VoteCount:            0,
		CreateTime:           now,
		PollType:             constants.PollTypePoll,
		HideResults:          constants.HideResultsOff,
		Anonymous:            false,
		StanceReasonRequired: constants.StanceReasonMust,
	}
	if err := repositories.VoteRepository.Create(sqls.DB(), vote); err != nil {
		t.Fatalf("create vote: %v", err)
	}

	opt1 := mustCreateVoteOption(t, vote.Id, "Option 1", now)

	form := StanceCreateForm{
		PollId:    vote.Id,
		OptionIds: []int64{opt1.Id},
		Reason:    "",
	}
	_, err := StanceService.CreateStance(user.Id, form)
	if err == nil {
		t.Fatalf("expected error when reason required but reason empty")
	}
}

func TestStanceService_CreateStance_UpdateExisting(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user.Id, now+86400, now)
	opt1 := mustCreateVoteOption(t, vote.Id, "Option 1", now)
	opt2 := mustCreateVoteOption(t, vote.Id, "Option 2", now)

	form1 := StanceCreateForm{
		PollId:    vote.Id,
		OptionIds: []int64{opt1.Id},
		Reason:    "first stance",
	}
	stance1, err := StanceService.CreateStance(user.Id, form1)
	if err != nil {
		t.Fatalf("create first stance: %v", err)
	}

	form2 := StanceCreateForm{
		PollId:    vote.Id,
		OptionIds: []int64{opt2.Id},
		Reason:    "updated stance",
	}
	stance2, err := StanceService.CreateStance(user.Id, form2)
	if err != nil {
		t.Fatalf("create second stance: %v", err)
	}

	if stance2.Id <= stance1.Id {
		t.Errorf("expected stance2.Id > stance1.Id")
	}

	stance1Check := StanceService.Get(stance1.Id)
	if stance1Check.Latest {
		t.Fatalf("expected stance1 to not be latest after update")
	}

	stance2Check := StanceService.Get(stance2.Id)
	if !stance2Check.Latest {
		t.Fatalf("expected stance2 to be latest")
	}

	choices := StanceService.GetStanceChoices(stance2.Id)
	if len(choices) != 1 {
		t.Fatalf("expected 1 choice for updated stance, got %d", len(choices))
	}
	if choices[0].PollOptionId != opt2.Id {
		t.Errorf("expected choice to be for option 2")
	}

	voteCheck := VoteService.Get(vote.Id)
	if voteCheck.VoteCount != 1 {
		t.Errorf("expected vote count 1, got %d", voteCheck.VoteCount)
	}

	opt1Check := VoteOptionService.Get(opt1.Id)
	if opt1Check.VoteCount != 0 {
		t.Errorf("expected option 1 vote count 0, got %d", opt1Check.VoteCount)
	}

	opt2Check := VoteOptionService.Get(opt2.Id)
	if opt2Check.VoteCount != 1 {
		t.Errorf("expected option 2 vote count 1, got %d", opt2Check.VoteCount)
	}
}

func TestStanceService_RevokeStance(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user.Id, now+86400, now)
	opt1 := mustCreateVoteOption(t, vote.Id, "Option 1", now)

	form := StanceCreateForm{
		PollId:    vote.Id,
		OptionIds: []int64{opt1.Id},
		Reason:    "test reason",
	}
	stance, err := StanceService.CreateStance(user.Id, form)
	if err != nil {
		t.Fatalf("create stance: %v", err)
	}

	err = StanceService.RevokeStance(user.Id, stance.Id)
	if err != nil {
		t.Fatalf("revoke stance: %v", err)
	}

	stanceCheck := StanceService.Get(stance.Id)
	if stanceCheck.Latest {
		t.Errorf("expected stance to not be latest after revoke")
	}
	if stanceCheck.RevokedAt == nil {
		t.Errorf("expected revokedAt to be set")
	}

	voteCheck := VoteService.Get(vote.Id)
	if voteCheck.VoteCount != 0 {
		t.Errorf("expected vote count 0, got %d", voteCheck.VoteCount)
	}
}

func TestOutcomeService_CreateOutcome(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user.Id, now+86400, now)

	form := OutcomeCreateForm{
		PollId:          vote.Id,
		Statement:       "test outcome statement",
		StatementFormat: "text",
	}
	outcome, err := OutcomeService.CreateOutcome(user.Id, form)
	if err != nil {
		t.Fatalf("create outcome: %v", err)
	}
	if outcome == nil {
		t.Fatalf("expected outcome to be created")
	}
	if outcome.PollId != vote.Id {
		t.Errorf("expected pollId %d, got %d", vote.Id, outcome.PollId)
	}
	if outcome.AuthorId != user.Id {
		t.Errorf("expected authorId %d, got %d", user.Id, outcome.AuthorId)
	}
	if outcome.Statement != "test outcome statement" {
		t.Errorf("expected statement 'test outcome statement', got '%s'", outcome.Statement)
	}
	if !outcome.Latest {
		t.Errorf("expected latest true")
	}

	outcomeCheck := OutcomeService.GetLatestByPollId(vote.Id)
	if outcomeCheck == nil {
		t.Fatalf("expected outcome to be found")
	}
	if outcomeCheck.Id != outcome.Id {
		t.Errorf("expected same outcome")
	}
}

func TestOutcomeService_CreateOutcome_NonOwner(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user1 := mustCreateVoteUser(t, now)
	user2 := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user1.Id, now+86400, now)

	form := OutcomeCreateForm{
		PollId:          vote.Id,
		Statement:       "test outcome statement",
		StatementFormat: "text",
	}
	_, err := OutcomeService.CreateOutcome(user2.Id, form)
	if err == nil {
		t.Fatalf("expected error for non-owner")
	}
}

func TestOutcomeService_UpdateOutcome(t *testing.T) {
	_ = setupVoteTestDB(t)
	now := time.Date(2025, 1, 2, 9, 0, 0, 0, time.UTC).UnixMilli()
	user := mustCreateVoteUser(t, now)
	vote := mustCreateVote(t, user.Id, now+86400, now)

	createForm := OutcomeCreateForm{
		PollId:          vote.Id,
		Statement:       "initial statement",
		StatementFormat: "text",
	}
	outcome, err := OutcomeService.CreateOutcome(user.Id, createForm)
	if err != nil {
		t.Fatalf("create outcome: %v", err)
	}

	updateForm := OutcomeCreateForm{
		PollId:          vote.Id,
		Statement:       "updated statement",
		StatementFormat: "text",
	}
	updatedOutcome, err := OutcomeService.UpdateOutcome(user.Id, outcome.Id, updateForm)
	if err != nil {
		t.Fatalf("update outcome: %v", err)
	}
	if updatedOutcome.Statement != "updated statement" {
		t.Errorf("expected statement 'updated statement', got '%s'", updatedOutcome.Statement)
	}
}
