package capabilityhost

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	"lina-core/internal/service/cachecoord"
	"lina-core/internal/service/coordination"
	"lina-core/pkg/plugin/capability/authcap/machinecoord"
)

func TestMachineCoordinationRejectsUnboundCalls(t *testing.T) {
	service := newMachineCoordinationAdapter(
		cachecoord.NewStaticTopology(false),
		nil,
		cachecoord.New(cachecoord.NewStaticTopology(false)),
	)
	if err := service.Configure(context.Background(), time.Second); err == nil {
		t.Fatal("expected unbound coordination configuration to fail")
	}
	if _, err := service.CurrentRevision(context.Background(), 1); err == nil {
		t.Fatal("expected unbound revision read to fail")
	}
}

func TestMachineCoordinationIsolatesPluginAndTenantRevisions(t *testing.T) {
	ctx := context.Background()
	cacheCoord := cachecoord.New(cachecoord.NewStaticTopology(false))
	base := newMachineCoordinationAdapter(cachecoord.NewStaticTopology(false), nil, cacheCoord)
	pluginA := base.forPlugin("plugin-a")
	pluginB := base.forPlugin("plugin-b")

	for _, service := range []machinecoord.Service{pluginA, pluginB} {
		if err := service.Configure(ctx, 3*time.Second); err != nil {
			t.Fatalf("configure machine coordination: %v", err)
		}
	}
	if revision, err := pluginA.MarkChanged(ctx, 7, machinecoord.ChangeReasonCredential); err != nil || revision != 1 {
		t.Fatalf("mark plugin-a tenant revision: revision=%d err=%v", revision, err)
	}
	if revision, err := pluginA.MarkChanged(ctx, 7, machinecoord.ChangeReasonPolicy); err != nil || revision != 2 {
		t.Fatalf("advance plugin-a tenant revision: revision=%d err=%v", revision, err)
	}
	if revision, err := pluginA.CurrentRevision(ctx, 8); err != nil || revision != 1 {
		t.Fatalf("expected tenant isolation, revision=%d err=%v", revision, err)
	}
	if revision, err := pluginB.CurrentRevision(ctx, 7); err != nil || revision != 1 {
		t.Fatalf("expected plugin isolation, revision=%d err=%v", revision, err)
	}
	if revision, err := pluginA.CurrentRevision(ctx, 7); err != nil || revision != 2 {
		t.Fatalf("read plugin-a tenant revision: revision=%d err=%v", revision, err)
	}

	snapshot, err := cacheCoord.Snapshot(ctx)
	if err != nil {
		t.Fatalf("read cache coordination snapshot: %v", err)
	}
	found := false
	for _, item := range snapshot {
		if item.Domain != machineAccessDomain {
			continue
		}
		found = true
		if item.ConsistencyModel != cachecoord.ConsistencyLocalOnly ||
			item.FailureStrategy != cachecoord.FailureStrategyFailClosed ||
			item.MaxStale != 3*time.Second {
			t.Fatalf("unexpected machine-access consistency metadata: %+v", item)
		}
	}
	if !found {
		t.Fatal("expected machine-access domain in coordination snapshot")
	}
}

func TestMachineCoordinationConsumesSharedReplayOnce(t *testing.T) {
	ctx := context.Background()
	coordinationSvc := coordination.NewMemory(coordination.NewKeyBuilder("test", "machine", "shared"))
	topology := cachecoord.NewStaticTopology(true)
	service := newMachineCoordinationAdapter(
		topology,
		coordinationSvc,
		cachecoord.NewWithCoordination(topology, coordinationSvc),
	).forPlugin("plugin-a")

	digestBytes := sha256.Sum256([]byte("ak-1\x00nonce-1"))
	digest := hex.EncodeToString(digestBytes[:])
	accepted, err := service.ConsumeSharedReplay(ctx, 9, digest, time.Minute)
	if err != nil || !accepted {
		t.Fatalf("consume first shared replay key: accepted=%v err=%v", accepted, err)
	}
	accepted, err = service.ConsumeSharedReplay(ctx, 9, digest, time.Minute)
	if err != nil {
		t.Fatalf("consume duplicate shared replay key: %v", err)
	}
	if accepted {
		t.Fatal("expected duplicate shared replay key to be rejected")
	}

	otherTenantAccepted, err := service.ConsumeSharedReplay(ctx, 10, digest, time.Minute)
	if err != nil || !otherTenantAccepted {
		t.Fatalf("expected tenant-isolated replay key: accepted=%v err=%v", otherTenantAccepted, err)
	}
}

func TestMachineCoordinationFailsClosedWithoutSharedBackend(t *testing.T) {
	topology := cachecoord.NewStaticTopology(true)
	service := newMachineCoordinationAdapter(
		topology,
		nil,
		cachecoord.New(topology),
	).forPlugin("plugin-a")
	digestBytes := sha256.Sum256([]byte("ak-1\x00nonce-1"))

	accepted, err := service.ConsumeSharedReplay(
		context.Background(),
		1,
		hex.EncodeToString(digestBytes[:]),
		time.Minute,
	)
	if err == nil || accepted || !strings.Contains(err.Error(), "unavailable") {
		t.Fatalf("expected unavailable shared backend to fail closed, accepted=%v err=%v", accepted, err)
	}
}

func TestMachineCoordinationSharesTenantRevisionAcrossNodes(t *testing.T) {
	ctx := context.Background()
	coordinationSvc := coordination.NewMemory(coordination.NewKeyBuilder("test", "machine", "revisions"))
	topology := cachecoord.NewStaticTopology(true)
	nodeA := newMachineCoordinationAdapter(
		topology,
		coordinationSvc,
		cachecoord.NewWithCoordination(topology, coordinationSvc),
	).forPlugin("plugin-a")
	nodeB := newMachineCoordinationAdapter(
		topology,
		coordinationSvc,
		cachecoord.NewWithCoordination(topology, coordinationSvc),
	).forPlugin("plugin-a")
	for _, node := range []machinecoord.Service{nodeA, nodeB} {
		if err := node.Configure(ctx, 2*time.Second); err != nil {
			t.Fatalf("configure clustered node: %v", err)
		}
	}
	written, err := nodeA.MarkChanged(ctx, 17, machinecoord.ChangeReasonPolicy)
	if err != nil {
		t.Fatalf("publish node A revision: %v", err)
	}
	observed, err := nodeB.CurrentRevision(ctx, 17)
	if err != nil || observed != written {
		t.Fatalf("node B did not observe shared revision: written=%d observed=%d err=%v", written, observed, err)
	}
	otherTenant, err := nodeB.CurrentRevision(ctx, 18)
	if err != nil || otherTenant == written {
		t.Fatalf("shared revision leaked across tenants: written=%d other=%d err=%v", written, otherTenant, err)
	}
}

func TestMachineCoordinationRevisionReadRecoversWithSharedBackend(t *testing.T) {
	ctx := context.Background()
	topology := cachecoord.NewStaticTopology(true)
	unavailable := newMachineCoordinationAdapter(
		topology,
		nil,
		cachecoord.New(topology),
	).forPlugin("plugin-a")
	if _, err := unavailable.CurrentRevision(ctx, 19); err == nil {
		t.Fatal("expected clustered revision read without backend to fail closed")
	}

	coordinationSvc := coordination.NewMemory(coordination.NewKeyBuilder("test", "machine", "recovery"))
	recovered := newMachineCoordinationAdapter(
		topology,
		coordinationSvc,
		cachecoord.NewWithCoordination(topology, coordinationSvc),
	).forPlugin("plugin-a")
	if err := recovered.Configure(ctx, time.Second); err != nil {
		t.Fatalf("configure recovered coordination: %v", err)
	}
	revision, err := recovered.MarkChanged(ctx, 19, machinecoord.ChangeReasonRecovery)
	if err != nil || revision <= 0 {
		t.Fatalf("recovered coordination did not publish revision: revision=%d err=%v", revision, err)
	}
}
