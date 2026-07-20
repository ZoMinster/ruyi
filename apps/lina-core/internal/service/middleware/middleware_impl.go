// middleware_impl.go implements request middleware for sessions, CORS,
// localization, and route-publication helpers. It relies on the injected auth,
// tenant, i18n, and role services so request paths share runtime state and do
// not create independent service graphs while handling HTTP traffic.

package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"lina-core/internal/model"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/authcap/authspi"
	"lina-core/pkg/plugin/pluginhost"
	"net/http"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/i18n/gi18n"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
)

const machineAuthorizationResultContextKey gctx.StrKey = "lina.machine.authorization.result"

// SessionStore returns the session store for external use (e.g., cleanup tasks).
func (s *serviceImpl) SessionStore() session.Store {
	return s.authSvc.SessionStore()
}

// PublishedRouteMiddlewares returns the published host middleware directory for plugin route composition.
func (s *serviceImpl) PublishedRouteMiddlewares() pluginhost.RouteMiddlewares {
	if s == nil {
		return nil
	}

	return pluginhost.NewRouteMiddlewares(
		ghttp.MiddlewareNeverDoneCtx,
		s.Response,
		s.CORS,
		s.RequestBodyLimit,
		s.Ctx,
		s.Auth,
		s.Tenancy,
		s.Permission,
	)
}

// Ctx injects business context into request.
func (s *serviceImpl) Ctx(r *ghttp.Request) {
	customCtx := &model.Context{}
	s.bizCtxSvc.Init(r, customCtx)
	locale := s.i18nSvc.ResolveRequestLocale(r)
	r.SetCtx(gi18n.WithLanguage(r.Context(), locale))
	s.bizCtxSvc.SetLocale(r.Context(), locale)
	r.Response.Header().Set("Content-Language", locale)
	r.Middleware.Next()
}

// CORS handles cross-origin requests.
func (s *serviceImpl) CORS(r *ghttp.Request) {
	r.Response.CORSDefault()
	r.Middleware.Next()
}

// Auth dispatches the exact Authorization scheme, preserving Bearer JWT user
// behavior while allowing registered providers to establish machine actors.
func (s *serviceImpl) Auth(r *ghttp.Request) {
	if r == nil {
		return
	}
	scheme, credential, ok := parseAuthorizationHeader(r.GetHeader("Authorization"))
	if !ok {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	if strings.EqualFold(scheme, "Bearer") {
		s.authenticateBearer(r, credential)
		return
	}
	if s.authProviders == nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	request := buildProviderAuthenticationRequest(r, scheme, credential)
	result, err := s.authProviders.Authenticate(r.Context(), scheme, request)
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	if result.Actor.Kind != authcap.ActorKindMachine ||
		strings.TrimSpace(result.Actor.SubjectID) == "" ||
		strings.TrimSpace(result.Actor.CredentialID) == "" ||
		result.Actor.TenantID < 0 {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}
	s.bizCtxSvc.SetActor(r.Context(), result.Actor)
	r.SetCtxVar(machineAuthorizationResultContextKey, result)
	r.Middleware.Next()
}

// authenticateBearer preserves the existing JWT authentication and business
// context injection path.
func (s *serviceImpl) authenticateBearer(r *ghttp.Request, tokenString string) {
	claims, err := s.authSvc.AuthenticateAccessToken(r.Context(), strings.TrimSpace(tokenString))
	if err != nil {
		r.Response.WriteStatus(http.StatusUnauthorized)
		return
	}

	// Inject user info into business context.
	s.bizCtxSvc.SetUser(r.Context(), claims.TokenId, claims.UserId, claims.Username, claims.Status, claims.ClientType.String())
	s.bizCtxSvc.SetTenant(r.Context(), claims.TenantId)
	s.bizCtxSvc.SetImpersonation(
		r.Context(),
		claims.ActingUserId,
		claims.TenantId,
		claims.IsImpersonation,
		claims.IsImpersonation,
	)
	r.Middleware.Next()
}

// parseAuthorizationHeader splits one scheme token from its non-empty payload.
func parseAuthorizationHeader(value string) (scheme string, credential string, ok bool) {
	trimmed := strings.TrimSpace(value)
	separator := strings.IndexAny(trimmed, " \t")
	if separator <= 0 {
		return "", "", false
	}
	scheme = strings.TrimSpace(trimmed[:separator])
	credential = strings.TrimSpace(trimmed[separator+1:])
	if _, err := authspi.NormalizeScheme(scheme); err != nil || credential == "" {
		return "", "", false
	}
	return scheme, credential, true
}

// buildProviderAuthenticationRequest snapshots transport inputs without
// exposing the mutable GoFrame request to authentication providers.
func buildProviderAuthenticationRequest(
	r *ghttp.Request,
	scheme string,
	credential string,
) authcap.AuthenticationRequest {
	query := make([]authcap.QueryParameter, 0)
	queryValues := r.URL.Query()
	queryKeys := make([]string, 0, len(queryValues))
	for key := range queryValues {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)
	for _, key := range queryKeys {
		for _, value := range queryValues[key] {
			query = append(query, authcap.QueryParameter{Key: key, Value: value})
		}
	}
	headerNames := make([]string, 0, len(r.Request.Header))
	for name := range r.Request.Header {
		headerNames = append(headerNames, name)
	}
	sort.Strings(headerNames)
	headers := make([]authcap.Header, 0, len(headerNames))
	for _, name := range headerNames {
		headers = append(headers, authcap.Header{
			Name:   name,
			Values: append([]string(nil), r.Request.Header.Values(name)...),
		})
	}
	bodyHash := sha256.Sum256(r.GetBody())
	escapedPath := r.URL.EscapedPath()
	if escapedPath == "" {
		escapedPath = "/"
	}
	return authcap.NewAuthenticationRequest(
		scheme,
		credential,
		r.Method,
		escapedPath,
		query,
		headers,
		hex.EncodeToString(bodyHash[:]),
		r.Request.RemoteAddr,
	)
}
