package services

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
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

func setupProposalVoteTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:proposal_vote_test_%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano())
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
		&models.Topic{},
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

func mustCreateProposalVoteUser(t *testing.T, now int64) *models.User {
	t.Helper()
	user := &models.User{
		Nickname:   fmt.Sprintf("user-%d", now),
		Status:     constants.StatusOk,
		CreateTime: now,
		UpdateTime: now,
	}
	if err := repositories.UserRepository.Create(sqls.DB(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func mustCreateProposalVoteTopic(t *testing.T, userId int64, now int64) *models.Topic {
	t.Helper()
	topic := &models.Topic{
		Type:       constants.TopicTypeTopic,
		CategoryId: 1,
		UserId:     userId,
		Title:      "proposal topic",
		Content:    "content",
		Status:     constants.StatusOk,
		CreateTime: now,
	}
	if err := repositories.TopicRepository.Create(sqls.DB(), topic); err != nil {
		t.Fatalf("create topic: %v", err)
	}
	return topic
}

func mustCreateProposalVote(t *testing.T, topicId, userId int64, form *req.VoteDTO, now int64) *models.Vote {
	t.Helper()
	var vote *models.Vote
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		var innerErr error
		vote, innerErr = VoteService.CreateWithOptionsTx(ctx, topicId, userId, form, now)
		return innerErr
	})
	if err != nil {
		t.Fatalf("create vote: %v", err)
	}
	return vote
}

func TestVoteService_CastRejectsProposalVote(t *testing.T) {
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
	if len(options) < 1 {
		t.Fatal("expected vote options")
	}

	err := VoteService.Cast(voter.Id, req.VoteCastReq{VoteId: vote.Id, OptionIds: []int64{options[0].Id}})
	if err == nil {
		t.Fatal("expected proposal vote cast to be rejected")
	}
}
