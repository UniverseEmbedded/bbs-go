package api

import (
	"bbs-go/internal/controllers/render"
	"bbs-go/internal/pkg/common"
	"bbs-go/internal/pkg/errs"
	"bbs-go/internal/services"

	"github.com/kataras/iris/v12"
	"github.com/mlogclub/simple/web"
)

type OutcomeController struct {
	Ctx iris.Context
}

func (c *OutcomeController) GetBy(outcomeId int64) *web.JsonResult {
	outcome := services.OutcomeService.Get(outcomeId)
	if outcome == nil {
		return web.JsonErrorMsg("outcome not found")
	}
	return web.JsonData(render.BuildOutcome(c.Ctx, outcome))
}

func (c *OutcomeController) PostCreate() *web.JsonResult {
	user := common.GetCurrentUser(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin())
	}

	var form services.OutcomeCreateForm
	if err := c.Ctx.ReadJSON(&form); err != nil {
		return web.JsonError(err)
	}

	outcome, err := services.OutcomeService.CreateOutcome(user.Id, form)
	if err != nil {
		return web.JsonError(err)
	}

	return web.JsonData(render.BuildOutcome(c.Ctx, outcome))
}

func (c *OutcomeController) PostUpdateBy(outcomeId int64) *web.JsonResult {
	user := common.GetCurrentUser(c.Ctx)
	if user == nil {
		return web.JsonError(errs.NotLogin())
	}

	var form services.OutcomeCreateForm
	if err := c.Ctx.ReadJSON(&form); err != nil {
		return web.JsonError(err)
	}

	outcome, err := services.OutcomeService.UpdateOutcome(user.Id, outcomeId, form)
	if err != nil {
		return web.JsonError(err)
	}

	return web.JsonData(render.BuildOutcome(c.Ctx, outcome))
}

func (c *OutcomeController) GetPoll() *web.JsonResult {
	pollId := common.GetID(c.Ctx, "pollId")
	if pollId <= 0 {
		return web.JsonErrorMsg("pollId is required")
	}

	outcome := services.OutcomeService.GetLatestByPollId(pollId)
	if outcome == nil {
		return web.JsonData(nil)
	}
	return web.JsonData(render.BuildOutcome(c.Ctx, outcome))
}
