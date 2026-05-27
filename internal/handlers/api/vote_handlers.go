package api

import (
	"bbs-go/internal/handlers/render"
	"bbs-go/internal/models/req"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"

	"bbs-go/internal/pkg/ginx"
)

func VoteDetail(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	voteId := id

	vote := services.VoteService.Get(voteId)
	if vote == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("vote not found"))
		return
	}
	ginx.WriteJSON(ctx, render.BuildVote(ctx, vote))
}

func VoteCast(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}

	var form req.VoteCastReq
	if err := ginx.BindJSON(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	err := services.VoteService.Cast(user.Id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}

	vote := services.VoteService.Get(form.VoteId)
	ginx.WriteJSON(ctx, render.BuildVote(ctx, vote))
}

func StanceCreate(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	var form req.StanceCreateReq
	if err := ginx.BindJSON(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	stance, err := services.StanceService.CreateStance(user.Id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, render.BuildStance(ctx, stance))
}

func StanceRevoke(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	stanceId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	if err := services.StanceService.RevokeStance(user.Id, stanceId); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, nil)
}

func StanceLatest(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	pollId, err := strconv.ParseInt(ctx.Query("pollId"), 10, 64)
	if err != nil || pollId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("pollId is required"))
		return
	}
	stance := services.StanceService.GetLatestByParticipant(pollId, user.Id)
	if stance == nil {
		ginx.WriteJSON(ctx, nil)
		return
	}
	ginx.WriteJSON(ctx, render.BuildStance(ctx, stance))
}

func StanceOptions(ctx *gin.Context) {
	pollId, err := strconv.ParseInt(ctx.Query("pollId"), 10, 64)
	if err != nil || pollId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("pollId is required"))
		return
	}
	vote := services.VoteService.Get(pollId)
	if vote == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("vote not found"))
		return
	}
	currentUserId := common.GetCurrentUserID(ctx)
	if !services.StanceService.CanViewResults(vote, currentUserId) {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("results not available"))
		return
	}
	stances := services.StanceService.FindByPollId(pollId)
	ginx.WriteJSON(ctx, render.BuildStanceList(ctx, stances))
}

func OutcomeCreate(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	var form req.OutcomeCreateReq
	if err := ginx.BindJSON(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	outcome, err := services.OutcomeService.CreateOutcome(user.Id, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, render.BuildOutcome(ctx, outcome))
}

func OutcomeUpdate(ctx *gin.Context) {
	user := common.GetCurrentUser(ctx)
	if user == nil {
		ginx.WriteJSON(ctx, errs.NotLogin())
		return
	}
	outcomeId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	var form req.OutcomeCreateReq
	if err := ginx.BindJSON(ctx, &form); err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	outcome, err := services.OutcomeService.UpdateOutcome(user.Id, outcomeId, form)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	ginx.WriteJSON(ctx, render.BuildOutcome(ctx, outcome))
}

func OutcomeDetail(ctx *gin.Context) {
	outcomeId, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ginx.WriteJSON(ctx, err)
		return
	}
	outcome := services.OutcomeService.Get(outcomeId)
	if outcome == nil {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("outcome not found"))
		return
	}
	ginx.WriteJSON(ctx, render.BuildOutcome(ctx, outcome))
}

func OutcomePoll(ctx *gin.Context) {
	pollId, err := strconv.ParseInt(ctx.Query("pollId"), 10, 64)
	if err != nil || pollId <= 0 {
		ginx.WriteJSON(ctx, ginx.ErrorMessage("pollId is required"))
		return
	}
	outcome := services.OutcomeService.GetLatestByPollId(pollId)
	if outcome == nil {
		ginx.WriteJSON(ctx, nil)
		return
	}
	ginx.WriteJSON(ctx, render.BuildOutcome(ctx, outcome))
}
