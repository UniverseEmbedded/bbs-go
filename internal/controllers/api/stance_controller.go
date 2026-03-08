package api

import (
	"bbs-go/internal/controllers/render"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

type StanceController struct {
	Ctx iris.Context
}

func (c *StanceController) GetBy(stanceId int64) *web.JsonResult {
	stance := services.StanceService.Get(stanceId)
	if stance == nil {
		return web.JsonErrorMsg("stance not found")
	}
	return web.JsonData(render.BuildStance(c.Ctx, stance))
}

func (c *StanceController) PostCreate() *web.JsonResult {
	user := common.GetCurrentUser(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin())
	}

	var form services.StanceCreateForm
	if err := c.Ctx.ReadJSON(&form); err != nil {
		return web.JsonError(err)
	}

	stance, err := services.StanceService.CreateStance(user.Id, form)
	if err != nil {
		return web.JsonError(err)
	}

	return web.JsonData(render.BuildStance(c.Ctx, stance))
}

func (c *StanceController) PostRevokeBy(stanceId int64) *web.JsonResult {
	user := common.GetCurrentUser(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin())
	}

	err := services.StanceService.RevokeStance(user.Id, stanceId)
	if err != nil {
		return web.JsonError(err)
	}

	return web.JsonSuccess()
}

func (c *StanceController) GetLatest() *web.JsonResult {
	user := common.GetCurrentUser(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin())
	}

	pollId := common.GetID(c.Ctx, "pollId")
	if pollId <= 0 {
		return web.JsonErrorMsg("pollId is required")
	}

	stance := services.StanceService.GetLatestByParticipant(pollId, user.Id)
	if stance == nil {
		return web.JsonData(nil)
	}
	return web.JsonData(render.BuildStance(c.Ctx, stance))
}

func (c *StanceController) GetOptions() *web.JsonResult {
	pollId := common.GetID(c.Ctx, "pollId")
	if pollId <= 0 {
		return web.JsonErrorMsg("pollId is required")
	}

	vote := services.VoteService.Get(pollId)
	if vote == nil {
		return web.JsonErrorMsg("vote not found")
	}

	if !services.StanceService.CanViewResults(c.Ctx, vote) {
		return web.JsonErrorMsg("results not available")
	}

	stances := services.StanceService.FindByPollId(pollId)
	return web.JsonData(render.BuildStanceList(c.Ctx, stances))
}
