//go:build ignore

package actions

import "context"

// BadExecute calls action.OK without actx.Resolve.
func BadExecute(ctx context.Context, actx interface{}) error {
	action.OK(ctx, actx)
	return nil
}
