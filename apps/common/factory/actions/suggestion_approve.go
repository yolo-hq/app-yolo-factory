package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"
	"github.com/yolo-hq/yolo/core/write"

	"github.com/yolo-hq/app-yolo-factory/.yolo/fields"
	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// SuggestionApprove approves a pending suggestion.
//
// @policy CanApproveSuggestionPolicy
func SuggestionApprove(ctx context.Context, actx *action.Context, in inputs.ApproveSuggestionInput) error {
	_, err := sm.Suggestion.Approve(ctx, actx, actx.EntityID, write.Set{
		fields.Suggestion.ConvertedTaskID.When(in.PRDID != "").Value(in.PRDID),
	})
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("suggestion is not pending")
	}
	if err != nil {
		return fmt.Errorf("approve-suggestion: %w", err)
	}
	return nil
}
