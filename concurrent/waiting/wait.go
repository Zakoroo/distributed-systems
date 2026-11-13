// Package waiting lets a caller attach a policy around a blocking wait without
// making the waiting component depend on a transport or semaphore implementation.
package waiting

import "context"

type policyKey struct{}

// WithPolicy returns a context whose blocking waits use policy. A policy must
// execute wait synchronously and restore any released resources before returning.
func WithPolicy(ctx context.Context, policy func(wait func() error) error) context.Context {
	return context.WithValue(ctx, policyKey{}, policy)
}

// Wait executes a blocking wait through the context's policy, if any.
func Wait(ctx context.Context, wait func() error) error {
	policy, _ := ctx.Value(policyKey{}).(func(func() error) error)
	if policy == nil {
		return wait()
	}
	return policy(wait)
}
