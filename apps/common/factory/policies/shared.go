// Package policies defines EntityPolicies for Factory domain.
// Named Can{Action}{Entity}Policy — each policy maps 1:1 with an action.
// One file per policy.
package policies

// statusData is the shared policy data struct — requests only the "status" field.
type statusData struct {
	Status string `policy:"status"`
}
