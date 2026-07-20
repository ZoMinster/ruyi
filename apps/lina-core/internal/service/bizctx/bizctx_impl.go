// bizctx_impl.go implements request business-context storage and retrieval.
// It keeps the context key handling centralized so middleware, controllers,
// and services share one request-scoped identity, tenant, role, and permission
// snapshot without rebuilding context state downstream.

package bizctx

import (
	"context"
	"strconv"

	"lina-core/internal/model"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/bizctxcap"

	"github.com/gogf/gf/v2/net/ghttp"
)

// Init initializes and injects business context into request.
func (s *serviceImpl) Init(r *ghttp.Request, ctx *model.Context) {
	r.SetCtxVar(ContextKey, ctx)
}

// Get retrieves business context from context.
func (s *serviceImpl) Get(ctx context.Context) *model.Context {
	value := ctx.Value(ContextKey)
	if value == nil {
		return nil
	}
	if localCtx, ok := value.(*model.Context); ok {
		return localCtx
	}
	return nil
}

// Current returns the plugin-visible read-only projection of the current
// business context.
func (s *serviceImpl) Current(ctx context.Context) bizctxcap.CurrentContext {
	if c := s.Get(ctx); c != nil {
		return bizctxcap.CurrentContext{
			Actor:                c.Actor,
			TokenID:              c.TokenId,
			UserID:               c.UserId,
			Username:             c.Username,
			TenantID:             c.TenantId,
			ActingUserID:         c.ActingUserId,
			ActingAsTenant:       c.ActingAsTenant,
			IsImpersonation:      c.IsImpersonation,
			DataScope:            c.DataScope,
			DataScopeUnsupported: c.DataScopeUnsupported,
			UnsupportedDataScope: c.UnsupportedDataScope,
			PlatformBypass: c.TenantId == 0 &&
				c.DataScope == 1 &&
				!c.DataScopeUnsupported &&
				!c.ActingAsTenant &&
				!c.IsImpersonation,
		}
	}
	return bizctxcap.CurrentFromContext(ctx)
}

// SetLocale sets locale info into business context.
func (s *serviceImpl) SetLocale(ctx context.Context, locale string) {
	if c := s.Get(ctx); c != nil {
		c.Locale = locale
	}
}

// SetUser sets user info into business context.
func (s *serviceImpl) SetUser(ctx context.Context, tokenId string, userId int, username string, status int, clientType string) {
	if c := s.Get(ctx); c != nil {
		c.Actor = authcap.Actor{
			Kind:         authcap.ActorKindUser,
			SubjectID:    strconv.Itoa(userId),
			CredentialID: tokenId,
			TenantID:     c.TenantId,
		}
		c.TokenId = tokenId
		c.UserId = userId
		c.Username = username
		c.Status = status
		c.ClientType = clientType
	}
}

// SetActor records a trusted actor and clears incompatible user authorization
// state when the actor is a machine.
func (s *serviceImpl) SetActor(ctx context.Context, actor authcap.Actor) {
	if c := s.Get(ctx); c != nil {
		c.Actor = actor
		c.TenantId = actor.TenantID
		if actor.Kind != authcap.ActorKindMachine {
			return
		}
		c.TokenId = ""
		c.UserId = 0
		c.Username = ""
		c.Status = 0
		c.ClientType = ""
		c.ActingAsTenant = false
		c.ActingUserId = 0
		c.IsImpersonation = false
		c.DataScope = 0
		c.DataScopeUnsupported = false
		c.UnsupportedDataScope = 0
	}
}

// SetTenant sets tenant info into business context.
func (s *serviceImpl) SetTenant(ctx context.Context, tenantId int) {
	if c := s.Get(ctx); c != nil {
		c.TenantId = tenantId
		c.Actor.TenantID = tenantId
	}
}

// SetImpersonation sets platform impersonation info into business context.
func (s *serviceImpl) SetImpersonation(ctx context.Context, actingUserId int, tenantId int, actingAsTenant bool, isImpersonation bool) {
	if c := s.Get(ctx); c != nil {
		if c.Actor.Kind == authcap.ActorKindMachine {
			return
		}
		c.ActingUserId = actingUserId
		c.TenantId = tenantId
		c.ActingAsTenant = actingAsTenant
		c.IsImpersonation = isImpersonation
	}
}

// SetUserAccess sets cached access-snapshot fields into business context.
func (s *serviceImpl) SetUserAccess(ctx context.Context, dataScope int, dataScopeUnsupported bool, unsupportedDataScope int) {
	if c := s.Get(ctx); c != nil {
		if c.Actor.Kind == authcap.ActorKindMachine {
			return
		}
		c.DataScope = dataScope
		c.DataScopeUnsupported = dataScopeUnsupported
		c.UnsupportedDataScope = unsupportedDataScope
	}
}
