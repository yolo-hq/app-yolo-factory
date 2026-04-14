package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/yolo-hq/yolo/core/action"

	"github.com/yolo-hq/app-yolo-factory/.yolo/sm"
	"github.com/yolo-hq/app-yolo-factory/apps/common/factory/inputs"
)

// SuggestionReject rejects a pending suggestion with a reason.
//
// @policy CanRejectSuggestionPolicy
func SuggestionReject(ctx context.Context, actx *action.Context, in inputs.RejectSuggestionInput) error {
	_ = in
	_, err := sm.Suggestion.Reject(ctx, actx, actx.EntityID, nil)
	if errors.Is(err, action.ErrStaleState) {
		return action.Fail("suggestion is not pending")
	}
	if err != nil {
		return fmt.Errorf("reject-suggestion: %w", err)
	}
	return nil
}
