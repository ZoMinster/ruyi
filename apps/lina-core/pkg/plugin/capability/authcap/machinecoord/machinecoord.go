// Package machinecoord defines host-owned coordination operations required by
// machine authentication providers. Implementations bind the calling plugin
// identity and never expose Redis, cachecoord, or host-internal key models.
package machinecoord

import (
	"context"
	"time"
)

// ChangeReason identifies a reviewable machine-access authority mutation.
type ChangeReason string

const (
	// ChangeReasonCredential records client or access-key authority changes.
	ChangeReasonCredential ChangeReason = "credential"
	// ChangeReasonPolicy records policy metadata or relation changes.
	ChangeReasonPolicy ChangeReason = "policy"
	// ChangeReasonRecovery records an explicit rebuild or recovery action.
	ChangeReasonRecovery ChangeReason = "recovery"
)

// Service coordinates one provider plugin's tenant-scoped machine access state.
type Service interface {
	// Configure declares the database-authoritative machine-access cache domain
	// and maximum accepted local staleness. Invalid or unbound calls fail closed.
	Configure(ctx context.Context, maxStaleness time.Duration) error
	// ClusterEnabled reports whether this process requires shared coordination.
	ClusterEnabled() bool
	// CurrentRevision returns the latest visible tenant machine-access revision.
	CurrentRevision(ctx context.Context, tenantID int) (int64, error)
	// MarkChanged publishes a tenant machine-access revision after authority commit.
	MarkChanged(ctx context.Context, tenantID int, reason ChangeReason) (int64, error)
	// ConsumeSharedReplay atomically stores one one-way replay-key digest in the
	// shared backend. It is valid only in cluster mode and fails closed otherwise.
	ConsumeSharedReplay(
		ctx context.Context,
		tenantID int,
		replayKeyDigest string,
		ttl time.Duration,
	) (accepted bool, err error)
}
