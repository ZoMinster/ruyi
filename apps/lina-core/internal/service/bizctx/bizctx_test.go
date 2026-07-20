// This file verifies tenant and impersonation fields in request business context.

package bizctx

import (
	"context"
	"testing"

	"lina-core/internal/model"
	"lina-core/pkg/plugin/capability/authcap"
)

// TestSetTenantAndImpersonationMutateBusinessContext verifies tenant fields are request-scoped.
func TestSetTenantAndImpersonationMutateBusinessContext(t *testing.T) {
	var (
		service     = New()
		businessCtx = &model.Context{}
		ctx         = context.WithValue(context.Background(), ContextKey, businessCtx)
	)

	service.SetTenant(ctx, 12)
	service.SetImpersonation(ctx, 1, 12, true, true)

	if businessCtx.TenantId != 12 || !businessCtx.ActingAsTenant || !businessCtx.IsImpersonation {
		t.Fatalf("expected tenant impersonation fields to be set, got %#v", businessCtx)
	}
	if businessCtx.ActingUserId != 1 {
		t.Fatalf("expected acting user field, got %#v", businessCtx)
	}
}

// TestSetMachineActorClearsUserAuthorizationState verifies a machine actor can
// never inherit user identity, session, impersonation, or role-scope fields.
func TestSetMachineActorClearsUserAuthorizationState(t *testing.T) {
	var (
		service     = New()
		businessCtx = &model.Context{
			TokenId:              "user-token",
			UserId:               9,
			Username:             "admin",
			ActingUserId:         7,
			ActingAsTenant:       true,
			IsImpersonation:      true,
			DataScope:            1,
			DataScopeUnsupported: true,
		}
		ctx = context.WithValue(context.Background(), ContextKey, businessCtx)
	)

	service.SetActor(ctx, authcap.Actor{
		Kind:         authcap.ActorKindMachine,
		SubjectID:    "machine-client-1",
		CredentialID: "AKIDEXAMPLE",
		TenantID:     42,
	})
	service.SetImpersonation(ctx, 1, 99, true, true)
	service.SetUserAccess(ctx, 1, false, 0)

	if businessCtx.Actor.Kind != authcap.ActorKindMachine || businessCtx.TenantId != 42 {
		t.Fatalf("expected trusted machine actor and tenant, got %#v", businessCtx)
	}
	if businessCtx.UserId != 0 || businessCtx.TokenId != "" || businessCtx.Username != "" {
		t.Fatalf("expected machine actor to have no user session fields, got %#v", businessCtx)
	}
	if businessCtx.ActingUserId != 0 || businessCtx.ActingAsTenant || businessCtx.IsImpersonation {
		t.Fatalf("expected machine actor to reject impersonation fields, got %#v", businessCtx)
	}
	if businessCtx.DataScope != 0 || businessCtx.DataScopeUnsupported {
		t.Fatalf("expected machine actor to have no user data scope, got %#v", businessCtx)
	}
}
