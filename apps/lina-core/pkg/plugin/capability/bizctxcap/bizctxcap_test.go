// This file verifies the published plugin business-context context helpers.

package bizctxcap

import (
	"context"
	"testing"

	"lina-core/pkg/plugin/capability/authcap"
)

// TestWithCurrentContextProvidesPluginVisibleSnapshot verifies source-plugin
// tests and adapters can inject context without importing host internal types.
func TestWithCurrentContextProvidesPluginVisibleSnapshot(t *testing.T) {
	ctx := WithCurrentContext(context.Background(), CurrentContext{
		UserID:          3,
		TenantID:        12,
		ActingUserID:    7,
		ActingAsTenant:  true,
		IsImpersonation: true,
		Permissions:     []string{"system:user:list"},
	})

	current := CurrentFromContext(ctx)
	if current.UserID != 3 || current.TenantID != 12 || current.ActingUserID != 7 {
		t.Fatalf("expected injected context snapshot, got %+v", current)
	}
	if !current.ActingAsTenant || !current.IsImpersonation || current.PlatformBypass {
		t.Fatalf("expected tenant impersonation snapshot, got %+v", current)
	}
	if len(current.Permissions) != 1 || current.Permissions[0] != "system:user:list" {
		t.Fatalf("expected cloned permissions, got %+v", current.Permissions)
	}
}

// TestWithCurrentContextNormalizesMachineActor verifies a host-projected
// machine context cannot carry forged user, permission, or bypass state.
func TestWithCurrentContextNormalizesMachineActor(t *testing.T) {
	current := CurrentFromContext(WithCurrentContext(context.Background(), CurrentContext{
		Actor: authcap.Actor{
			Kind:         authcap.ActorKindMachine,
			SubjectID:    "machine-client-1",
			CredentialID: "AKIDEXAMPLE",
			TenantID:     99,
		},
		TokenID:        "forged-token",
		UserID:         1,
		Username:       "forged-user",
		TenantID:       42,
		Permissions:    []string{"*:*:*"},
		DataScope:      1,
		IsSuperAdmin:   true,
		PlatformBypass: true,
	}))

	if current.Actor.TenantID != 42 || current.TenantID != 42 {
		t.Fatalf("expected host context tenant to win, got %+v", current)
	}
	if current.UserID != 0 || current.TokenID != "" || current.Username != "" {
		t.Fatalf("expected machine context to clear user fields, got %+v", current)
	}
	if len(current.Permissions) != 0 || current.DataScope != 0 || current.IsSuperAdmin || current.PlatformBypass {
		t.Fatalf("expected machine context to clear user authorization fields, got %+v", current)
	}
}

// TestWithCurrentContextMarksPlatformBypass verifies platform-scope helper
// semantics remain available without a public service implementation.
func TestWithCurrentContextMarksPlatformBypass(t *testing.T) {
	current := CurrentFromContext(WithCurrentContext(context.Background(), CurrentContext{TenantID: 0}))
	if !current.PlatformBypass {
		t.Fatalf("expected platform bypass for platform tenant, got %+v", current)
	}
}

// TestCurrentFromContextClonesPermissions verifies callers cannot mutate the
// stored context snapshot through a returned permission slice.
func TestCurrentFromContextClonesPermissions(t *testing.T) {
	ctx := WithCurrentContext(context.Background(), CurrentContext{Permissions: []string{"one"}})
	first := CurrentFromContext(ctx)
	first.Permissions[0] = "mutated"

	second := CurrentFromContext(ctx)
	if second.Permissions[0] != "one" {
		t.Fatalf("expected context permissions to remain immutable, got %+v", second.Permissions)
	}
}
