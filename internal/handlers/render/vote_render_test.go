package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/repositories"
	"bbs-go/internal/services"
	"fmt"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/gin-gonic/gin"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func setupVoteRenderTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:vote_render_test_%d?mode=memory&cache=shared&_fk=1", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "t_", SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Silent),
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
	t.Cleanup(func() { _ = sqlDB.Close() })
	sqls.SetDB(db)
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Vote{}, &models.VoteOption{}, &models.VoteRecord{}, &models.Stance{}, &models.StanceChoice{}, &models.Outcome{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	return db
}

func mustCreateRenderUser(t *testing.T, now int64) *models.User {
	t.Helper()
	user := &models.User{Nickname: fmt.Sprintf("render-%d", now), Status: constants.StatusOk, CreateTime: now, UpdateTime: now}
	if err := repositories.UserRepository.Create(sqls.DB(), user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func mustCreateRenderTopic(t *testing.T, userId, now int64) *models.Topic {
	t.Helper()
	topic := &models.Topic{Type: constants.TopicTypeTopic, CategoryId: 1, UserId: userId, Title: "topic", Content: "content", Status: constants.StatusOk, CreateTime: now}
	if err := repositories.TopicRepository.Create(sqls.DB(), topic); err != nil {
		t.Fatalf("create topic: %v", err)
	}
	return topic
}

func mustCreateRenderVote(t *testing.T, topicId, userId int64, form *req.VoteDTO, now int64) *models.Vote {
	t.Helper()
	var vote *models.Vote
	err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		var innerErr error
		vote, innerErr = services.VoteService.CreateWithOptionsTx(ctx, topicId, userId, form, now)
		return innerErr
	})
	if err != nil {
		t.Fatalf("create vote: %v", err)
	}
	return vote
}

func TestBuildVote_ProposalRespectsVisibilityAndOutcome(t *testing.T) {
	setupVoteRenderTestDB(t)
	now := time.Now().UnixMilli()
	owner := mustCreateRenderUser(t, now)
	viewer := mustCreateRenderUser(t, now+1)
	topic := mustCreateRenderTopic(t, owner.Id, now)
	vote := mustCreateRenderVote(t, topic.Id, owner.Id, &req.VoteDTO{
		PollType:             constants.PollTypeProposal,
		Type:                 constants.VoteTypeMultiple,
		Title:                "proposal",
		ExpiredAt:            now + int64(time.Hour/time.Millisecond),
		VoteNum:              2,
		HideResults:          constants.HideResultsUntilVote,
		StanceReasonRequired: constants.StanceReasonOptional,
		Options: []req.VoteOptionDTO{{Content: "A", Meaning: "agree"}, {Content: "B", Prompt: "consider"}},
	}, now)
	options := services.VoteOptionService.FindByVoteId(vote.Id)
	if len(options) != 2 {
		t.Fatalf("expected 2 options, got %d", len(options))
	}
	if _, err := services.StanceService.CreateStance(viewer.Id, req.StanceCreateReq{PollId: vote.Id, OptionIds: []int64{options[0].Id}, Reason: "because"}); err != nil {
		t.Fatalf("create stance: %v", err)
	}
	if _, err := services.OutcomeService.CreateOutcome(owner.Id, req.OutcomeCreateReq{PollId: vote.Id, Statement: "done", PollOptionId: &options[0].Id}); err != nil {
		t.Fatalf("create outcome: %v", err)
	}

	ctxAnon, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctxAnon.Request = httptest.NewRequest("GET", "/", nil)
	anonResp := BuildVote(ctxAnon, services.VoteService.Get(vote.Id))
	if anonResp.CanViewResults {
		t.Fatal("anonymous viewer should not see hidden results before voting")
	}
	if len(anonResp.Options) != 0 {
		t.Fatalf("expected hidden options for anonymous viewer, got %d", len(anonResp.Options))
	}
	if anonResp.Outcome == nil {
		t.Fatal("expected outcome to be returned")
	}

	ctxViewer, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctxViewer.Request = httptest.NewRequest("GET", "/", nil)
	common.SetCurrentUser(ctxViewer, viewer)
	viewerResp := BuildVote(ctxViewer, services.VoteService.Get(vote.Id))
	if !viewerResp.CanViewResults {
		t.Fatal("voter should see results after stance")
	}
	if !viewerResp.Voted || len(viewerResp.OptionIds) != 1 || viewerResp.OptionIds[0] != options[0].Id {
		t.Fatalf("unexpected viewer vote response: %+v", viewerResp)
	}
	if len(viewerResp.Options) != 2 {
		t.Fatalf("expected 2 rendered options, got %d", len(viewerResp.Options))
	}
	if viewerResp.Options[0].Meaning == "" {
		t.Fatal("expected proposal option meaning to be rendered")
	}
	if viewerResp.Outcome == nil || viewerResp.Outcome.PollOption == nil {
		t.Fatal("expected outcome poll option rendered")
	}
}
