// This file implements declarative permission enforcement for static host APIs.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/net/ghttp"

	i18nsvc "lina-core/internal/service/i18n"
	"lina-core/internal/service/role"
	"lina-core/pkg/bizerr"
	"lina-core/pkg/plugin/capability/authcap"
)

// Permission middleware constants define metadata tag names, wildcard grants,
// and the normalized JSON error envelope code.
const (
	staticPermissionMetaTag  = "permission"
	staticPermissionWildcard = "*:*:*"
	staticOperationMetaTag   = "operation"
	staticResourceMetaTag    = "resource"
	staticActionMetaTag      = "action"
	staticActorsMetaTag      = "actors"
)

// staticPermissionContextKey stores manually declared permissions for routes
// that are not bound from a g.Meta-backed DTO.
type staticPermissionContextKey string

const manualStaticPermissionContextKey staticPermissionContextKey = "lina.static.permission"

// RequirePermission declares static permission requirements for manually registered routes.
func (s *serviceImpl) RequirePermission(permissions ...string) ghttp.HandlerFunc {
	normalizedPermissions := normalizePermissionList(strings.Join(permissions, ","))
	return func(r *ghttp.Request) {
		if len(normalizedPermissions) > 0 {
			r.SetCtxVar(manualStaticPermissionContextKey, strings.Join(normalizedPermissions, ","))
		}
		r.Middleware.Next()
	}
}

// Permission enforces declarative permission requirements declared on static host API handlers.
func (s *serviceImpl) Permission(r *ghttp.Request) {
	if r == nil {
		return
	}

	businessCtx := s.bizCtxSvc.Get(r.Context())
	if businessCtx != nil && businessCtx.Actor.Kind == authcap.ActorKindMachine {
		s.authorizeMachineRoute(r, businessCtx.Actor)
		return
	}
	if !routeAllowsActor(r, authcap.ActorKindUser) {
		writePermissionError(r, s.i18nSvc, http.StatusForbidden, bizerr.NewCode(CodeMiddlewareHTTPForbidden))
		return
	}

	requiredPermissions := extractDeclaredPermissions(r)
	if len(requiredPermissions) == 0 {
		// Build-time audit tests ensure protected static APIs declare permissions.
		// Middleware therefore treats "no metadata" as "no extra permission gate".
		r.Middleware.Next()
		return
	}

	if businessCtx == nil || businessCtx.UserId <= 0 {
		writePermissionError(
			r,
			s.i18nSvc,
			http.StatusUnauthorized,
			bizerr.NewCode(CodeMiddlewarePermissionCurrentUserMissing),
		)
		return
	}

	accessContext, err := s.roleSvc.GetUserAccessContext(r.Context(), businessCtx.UserId)
	if err != nil {
		writePermissionError(
			r,
			s.i18nSvc,
			http.StatusInternalServerError,
			bizerr.WrapCode(err, CodeMiddlewarePermissionContextLoadFailed),
		)
		return
	}
	s.bizCtxSvc.SetUserAccess(
		r.Context(),
		int(accessContext.DataScope),
		accessContext.DataScopeUnsupported,
		accessContext.UnsupportedDataScope,
	)
	if s.tenantSvc != nil && s.tenantSvc.PlatformBypass(r.Context()) {
		s.bizCtxSvc.SetTenant(r.Context(), 0)
	}
	if hasRequiredPermissions(accessContext, requiredPermissions) {
		r.Middleware.Next()
		return
	}

	writePermissionError(
		r,
		s.i18nSvc,
		http.StatusForbidden,
		bizerr.NewCode(
			CodeMiddlewarePermissionDeniedRequired,
			bizerr.P("permissions", strings.Join(requiredPermissions, ", ")),
		),
	)
}

// authorizeMachineRoute enforces explicit actor opt-in plus exact operation and
// resource-wide access grants. User permission tags are intentionally ignored.
func (s *serviceImpl) authorizeMachineRoute(r *ghttp.Request, actor authcap.Actor) {
	declaration, err := extractRouteAuthorization(r)
	if err != nil || !declaration.AllowsActor(authcap.ActorKindMachine) {
		writePermissionError(r, s.i18nSvc, http.StatusForbidden, bizerr.NewCode(CodeMiddlewareHTTPForbidden))
		return
	}
	value := r.GetCtxVar(machineAuthorizationResultContextKey).Val()
	result, ok := value.(authcap.AuthenticationResult)
	if !ok || result.Actor != actor || !result.Authorization.Allows(authcap.AuthorizationRequest{
		Operation: declaration.Operation,
		Resource:  declaration.Resource,
		Access:    declaration.Access,
	}) {
		writePermissionError(r, s.i18nSvc, http.StatusForbidden, bizerr.NewCode(CodeMiddlewareHTTPForbidden))
		return
	}
	r.Middleware.Next()
}

// routeAllowsActor parses the current DTO route declaration and defaults to
// user-only when actors metadata is absent.
func routeAllowsActor(r *ghttp.Request, actor authcap.ActorKind) bool {
	declaration, err := extractRouteAuthorization(r)
	return err == nil && declaration.AllowsActor(actor)
}

// extractRouteAuthorization reads the current strict handler's machine route
// metadata. Raw routes have no tags and therefore parse as user-only.
func extractRouteAuthorization(r *ghttp.Request) (authcap.RouteAuthorization, error) {
	var (
		operation string
		resource  string
		action    string
		actors    string
	)
	if r != nil {
		if handler := r.GetServeHandler(); handler != nil {
			operation = handler.GetMetaTag(staticOperationMetaTag)
			resource = handler.GetMetaTag(staticResourceMetaTag)
			action = handler.GetMetaTag(staticActionMetaTag)
			actors = handler.GetMetaTag(staticActorsMetaTag)
		}
	}
	method := ""
	path := ""
	if r != nil {
		method = r.Method
		path = r.URL.Path
	}
	return authcap.ParseRouteAuthorization(
		authcap.RouteOwnerKindHost,
		"request",
		method,
		path,
		operation,
		resource,
		action,
		actors,
	)
}

// extractDeclaredPermissions reads the permission metadata declared on the
// current request DTO/handler and normalizes it into one deduplicated list.
func extractDeclaredPermissions(r *ghttp.Request) []string {
	if r == nil {
		return nil
	}
	handler := r.GetServeHandler()
	if handler != nil {
		if permissions := resolveDeclaredPermissions(handler.GetMetaTag(staticPermissionMetaTag)); len(permissions) > 0 {
			return permissions
		}
	}
	if permissions := normalizePermissionList(r.GetCtxVar(manualStaticPermissionContextKey).String()); len(permissions) > 0 {
		return permissions
	}
	return nil
}

// resolveDeclaredPermissions normalizes the canonical permission metadata tag.
func resolveDeclaredPermissions(permissionTag string) []string {
	return normalizePermissionList(permissionTag)
}

// normalizePermissionList trims, deduplicates, and preserves order for the
// comma-separated permission list declared in route metadata.
func normalizePermissionList(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var (
		parts  = strings.Split(raw, ",")
		result = make([]string, 0, len(parts))
		seen   = make(map[string]struct{}, len(parts))
	)
	for _, part := range parts {
		permission := strings.TrimSpace(part)
		if permission == "" {
			continue
		}
		if _, ok := seen[permission]; ok {
			continue
		}
		seen[permission] = struct{}{}
		result = append(result, permission)
	}
	return result
}

// hasRequiredPermissions applies the static-host permission semantics: super
// admin and wildcard bypass, otherwise every declared permission must be granted.
func hasRequiredPermissions(accessContext *role.UserAccessContext, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if accessContext == nil {
		return false
	}
	if accessContext.IsSuperAdmin {
		return true
	}

	granted := make(map[string]struct{}, len(accessContext.Permissions))
	for _, permission := range accessContext.Permissions {
		currentPermission := strings.TrimSpace(permission)
		if currentPermission == "" {
			continue
		}
		granted[currentPermission] = struct{}{}
	}
	if _, ok := granted[staticPermissionWildcard]; ok {
		return true
	}

	for _, permission := range required {
		if _, ok := granted[permission]; !ok {
			return false
		}
	}
	return true
}

// writePermissionError writes one JSON error payload and binds the error onto
// the request so upper layers can still observe the failure cause.
func writePermissionError(r *ghttp.Request, i18nSvc i18nsvc.Service, status int, err error) {
	if r == nil {
		return
	}

	message := ""
	if i18nSvc != nil {
		message = i18nSvc.LocalizeError(r.Context(), err)
	}
	if message == "" && err != nil {
		message = err.Error()
	}

	r.SetError(err)
	r.Response.WriteStatus(status)
	var code gcode.Code = gcode.CodeUnknown
	if messageErr, ok := bizerr.As(err); ok {
		code = messageErr.TypeCode()
	}
	response := runtimeHandlerResponse{
		Code:    code.Code(),
		Data:    nil,
		Message: message,
	}
	applyRuntimeErrorMetadata(&response, err)
	r.Response.WriteJson(response)
}
