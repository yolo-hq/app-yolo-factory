// Package policies defines EntityPolicies for Factory domain.
// Named Can{Action}{Entity}Policy — each policy maps 1:1 with an action.
// One file per policy.
//
// All policies embed policy.TypedData[T] where T declares entity fields
// via `field:"..."` tags. The framework loads the data before Evaluate()
// runs and each policy reads it via p.Data(actx).
package policies
