package migrations

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"

	"github.com/mlogclub/simple/sqls"
)

func migrate_proposal_vote_support() error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		tx := ctx.Tx

		if err := tx.Model(&models.Vote{}).
			Where("poll_type = '' OR poll_type IS NULL").
			Update("poll_type", constants.PollTypePoll).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Vote{}).
			Where("hide_results IS NULL OR hide_results = 0").
			Update("hide_results", constants.HideResultsOff).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.Vote{}).
			Where("stance_reason_required IS NULL").
			Update("stance_reason_required", constants.StanceReasonDisabled).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.VoteOption{}).
			Where("total_score IS NULL").
			Update("total_score", 0).Error; err != nil {
			return err
		}

		if err := tx.Model(&models.VoteOption{}).
			Where("voter_count IS NULL").
			Update("voter_count", 0).Error; err != nil {
			return err
		}

		return nil
	})
}
