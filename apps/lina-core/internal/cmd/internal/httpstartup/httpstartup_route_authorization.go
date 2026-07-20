// This file audits host, source-plugin, and dynamic-plugin route authorization
// metadata and publishes the startup-owned global machine route catalog.

package httpstartup

import (
	"context"
	"reflect"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gmeta"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/pluginhost"
)

const hostRouteAuthorizationOwnerID = "lina-core"

// syncHTTPRouteAuthorizations publishes static and installed source-plugin route
// snapshots, then rebuilds installed dynamic-plugin snapshots from active releases.
func syncHTTPRouteAuthorizations(
	ctx context.Context,
	server *ghttp.Server,
	runtime *httpRuntime,
) error {
	if server == nil || runtime == nil || runtime.routeAuthorizations == nil || runtime.pluginSvc == nil {
		return gerror.New("HTTP route authorization startup dependencies are unavailable")
	}

	sourceBindings := runtime.pluginSvc.ListSourceRouteBindings()
	sourceRouteKeys := make(map[string]struct{}, len(sourceBindings))
	sourceRoutesByOwner := make(map[string][]authcap.RouteAuthorization)
	for _, binding := range sourceBindings {
		sourceRouteKeys[binding.Key()] = struct{}{}
		ownerKey := authcap.RouteAuthorizationOwnerKey(authcap.RouteOwnerKindSourcePlugin, binding.PluginID)
		sourceRoutesByOwner[ownerKey] = append(sourceRoutesByOwner[ownerKey], binding.Authorization)
	}

	hostRoutes, err := collectHostRouteAuthorizations(server, sourceRouteKeys)
	if err != nil {
		return err
	}
	hostOwner := authcap.RouteAuthorizationOwnerKey(authcap.RouteOwnerKindHost, hostRouteAuthorizationOwnerID)
	if err = runtime.routeAuthorizations.ReplaceOwner(hostOwner, hostRoutes); err != nil {
		return gerror.Wrap(err, "publish host route authorization catalog failed")
	}
	for ownerKey, routes := range sourceRoutesByOwner {
		if err = runtime.routeAuthorizations.ReplaceOwner(ownerKey, routes); err != nil {
			return gerror.Wrapf(err, "publish source route authorization catalog failed: %s", ownerKey)
		}
	}
	if err = runtime.pluginSvc.SyncDynamicRouteAuthorizations(ctx); err != nil {
		return gerror.Wrap(err, "publish dynamic route authorization catalog failed")
	}
	return nil
}

// collectHostRouteAuthorizations extracts request DTO metadata from strict host
// routes while excluding source-plugin routes already captured by the registrar.
func collectHostRouteAuthorizations(
	server *ghttp.Server,
	sourceRouteKeys map[string]struct{},
) ([]authcap.RouteAuthorization, error) {
	if server == nil {
		return nil, gerror.New("host route authorization server cannot be nil")
	}
	routes := make([]authcap.RouteAuthorization, 0)
	for _, route := range server.GetRoutes() {
		if route.Handler == nil || !route.Handler.Info.IsStrictRoute {
			continue
		}
		key := normalizeHTTPRouteAuthorizationKey(route.Method, route.Route)
		if _, sourceOwned := sourceRouteKeys[key]; sourceOwned {
			continue
		}
		item, err := parseHandlerRouteAuthorization(
			route.Handler.Info.Value.Interface(),
			authcap.RouteOwnerKindHost,
			hostRouteAuthorizationOwnerID,
			route.Method,
			route.Route,
		)
		if err != nil {
			return nil, gerror.Wrapf(err, "host route authorization metadata is invalid: %s", key)
		}
		routes = append(routes, item)
	}
	return routes, nil
}

// parseHandlerRouteAuthorization reads custom machine authorization tags from
// a standard GoFrame `(context.Context, *Req) (*Res, error)` handler.
func parseHandlerRouteAuthorization(
	handler interface{},
	ownerKind authcap.RouteOwnerKind,
	ownerID string,
	method string,
	path string,
) (authcap.RouteAuthorization, error) {
	reqObject := routeAuthorizationRequestObject(handler)
	return authcap.ParseRouteAuthorization(
		ownerKind,
		ownerID,
		method,
		path,
		readHTTPRouteAuthorizationTag(reqObject, "operation"),
		readHTTPRouteAuthorizationTag(reqObject, "resource"),
		readHTTPRouteAuthorizationTag(reqObject, "action"),
		readHTTPRouteAuthorizationTag(reqObject, "actors"),
	)
}

func routeAuthorizationRequestObject(handler interface{}) interface{} {
	handlerType := reflect.TypeOf(handler)
	if handlerType == nil || handlerType.Kind() != reflect.Func || handlerType.NumIn() != 2 {
		return nil
	}
	requestType := handlerType.In(1)
	if requestType.Kind() != reflect.Pointer || requestType.Elem().Kind() != reflect.Struct {
		return nil
	}
	return reflect.New(requestType.Elem()).Interface()
}

func readHTTPRouteAuthorizationTag(request interface{}, name string) string {
	if request == nil {
		return ""
	}
	return gmeta.Get(request, name).String()
}

func normalizeHTTPRouteAuthorizationKey(method string, path string) string {
	return pluginhost.SourceRouteBinding{Method: method, Path: path}.Key()
}
