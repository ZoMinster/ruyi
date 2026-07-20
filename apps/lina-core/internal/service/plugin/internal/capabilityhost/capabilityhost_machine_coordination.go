// This file adapts host cluster, cache revision, and coordination services to
// the plugin-bound machine authentication coordination contract.

package capabilityhost

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/cluster"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/plugin/capability/authcap/machinecoord"
)

const machineAccessDomain cachecoord.Domain = "machine-access"

type pluginBoundMachineCoordination interface {
	machinecoord.Service
	forPlugin(pluginID string) machinecoord.Service
}

type machineCoordinationAdapter struct {
	pluginID     string
	cluster      cluster.Service
	coordination coordination.Service
	cacheCoord   cachecoord.Service
}

var (
	_ machinecoord.Service           = (*machineCoordinationAdapter)(nil)
	_ pluginBoundMachineCoordination = (*machineCoordinationAdapter)(nil)
)

func newMachineCoordinationAdapter(
	clusterSvc cluster.Service,
	coordinationSvc coordination.Service,
	cacheCoordSvc cachecoord.Service,
) *machineCoordinationAdapter {
	return &machineCoordinationAdapter{
		cluster: clusterSvc, coordination: coordinationSvc, cacheCoord: cacheCoordSvc,
	}
}

func (s *machineCoordinationAdapter) forPlugin(pluginID string) machinecoord.Service {
	if s == nil {
		return nil
	}
	clone := *s
	clone.pluginID = strings.TrimSpace(pluginID)
	return &clone
}

// Configure registers the fixed machine-access authority and consistency contract.
func (s *machineCoordinationAdapter) Configure(_ context.Context, maxStaleness time.Duration) error {
	if err := s.validateBound(); err != nil {
		return err
	}
	if maxStaleness <= 0 {
		return gerror.New("machine access maximum staleness must be positive")
	}
	consistency := cachecoord.ConsistencyLocalOnly
	mechanism := "startup-shared process revision"
	if s.ClusterEnabled() {
		consistency = cachecoord.ConsistencySharedRevision
		mechanism = "shared revision and coordination event"
	}
	return s.cacheCoord.ConfigureDomain(cachecoord.DomainSpec{
		Domain:           machineAccessDomain,
		AuthoritySource:  "machine credential and access policy database tables",
		ConsistencyModel: consistency,
		MaxStale:         maxStaleness,
		SyncMechanism:    mechanism,
		FailureStrategy:  cachecoord.FailureStrategyFailClosed,
	})
}

// ClusterEnabled reports whether shared revision and replay state are mandatory.
func (s *machineCoordinationAdapter) ClusterEnabled() bool {
	return s != nil && s.cluster != nil && s.cluster.IsEnabled()
}

// CurrentRevision reads one plugin-bound tenant machine-access revision.
func (s *machineCoordinationAdapter) CurrentRevision(ctx context.Context, tenantID int) (int64, error) {
	if err := s.validateTenant(tenantID); err != nil {
		return 0, err
	}
	return s.cacheCoord.CurrentRevision(
		ctx,
		machineAccessDomain,
		cachecoord.ScopedScope(s.scope(), cachecoord.InvalidationScope{TenantID: cachecoord.TenantID(tenantID)}),
	)
}

// MarkChanged publishes one post-commit tenant revision.
func (s *machineCoordinationAdapter) MarkChanged(
	ctx context.Context,
	tenantID int,
	reason machinecoord.ChangeReason,
) (int64, error) {
	if err := s.validateTenant(tenantID); err != nil {
		return 0, err
	}
	if reason != machinecoord.ChangeReasonCredential && reason != machinecoord.ChangeReasonPolicy &&
		reason != machinecoord.ChangeReasonRecovery {
		return 0, gerror.New("machine access revision reason is invalid")
	}
	return s.cacheCoord.MarkTenantChanged(
		ctx,
		machineAccessDomain,
		s.scope(),
		cachecoord.InvalidationScope{TenantID: cachecoord.TenantID(tenantID)},
		cachecoord.ChangeReason(reason),
	)
}

// ConsumeSharedReplay atomically stores one digest in the cluster coordination backend.
func (s *machineCoordinationAdapter) ConsumeSharedReplay(
	ctx context.Context,
	tenantID int,
	replayKeyDigest string,
	ttl time.Duration,
) (bool, error) {
	if err := s.validateTenant(tenantID); err != nil {
		return false, err
	}
	if !s.ClusterEnabled() || s.coordination == nil || s.coordination.KV() == nil ||
		s.coordination.KeyBuilder() == nil {
		return false, gerror.New("shared machine replay coordination is unavailable")
	}
	replayKeyDigest = strings.TrimSpace(replayKeyDigest)
	decoded, err := hex.DecodeString(replayKeyDigest)
	if err != nil || len(decoded) != 32 || hex.EncodeToString(decoded) != replayKeyDigest {
		return false, gerror.New("machine replay key digest is invalid")
	}
	if ttl <= 0 {
		return false, gerror.New("machine replay ttl must be positive")
	}
	key, err := s.coordination.KeyBuilder().RawKVKey(
		"machine-auth-replay",
		s.pluginID,
		strconv.Itoa(tenantID),
		replayKeyDigest,
	)
	if err != nil {
		return false, err
	}
	return s.coordination.KV().SetNX(ctx, key, "1", ttl)
}

func (s *machineCoordinationAdapter) validateBound() error {
	if s == nil || strings.TrimSpace(s.pluginID) == "" || s.cacheCoord == nil {
		return gerror.New("machine access coordination is not plugin scoped")
	}
	return nil
}

func (s *machineCoordinationAdapter) validateTenant(tenantID int) error {
	if err := s.validateBound(); err != nil {
		return err
	}
	if tenantID < 0 {
		return gerror.New("machine access coordination tenant is invalid")
	}
	return nil
}

func (s *machineCoordinationAdapter) scope() cachecoord.Scope {
	return cachecoord.Scope("provider-" + s.pluginID)
}
