// This file projects the shared route authorization catalog into OpenAPI
// machine authentication requirements and stable authorization extensions.

package apidoc

import (
	"strings"

	"github.com/gogf/gf/v2/net/goai"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/pluginhost"
)

const (
	openAPIOperationExtension = "x-lina-operation"
	openAPIResourceExtension  = "x-lina-resource"
	openAPIActionExtension    = "x-lina-action"
	openAPIActorsExtension    = "x-lina-actors"
)

// projectMachineAuthentication applies machine security only to routes that
// the runtime catalog explicitly exposes to machine actors.
func (s *serviceImpl) projectMachineAuthentication(paths goai.Paths) {
	if s == nil || s.routeAuth == nil {
		return
	}
	for pathName, pathItem := range paths {
		operations := []struct {
			method    string
			operation *goai.Operation
		}{
			{method: "CONNECT", operation: pathItem.Connect},
			{method: "DELETE", operation: pathItem.Delete},
			{method: "GET", operation: pathItem.Get},
			{method: "HEAD", operation: pathItem.Head},
			{method: "OPTIONS", operation: pathItem.Options},
			{method: "PATCH", operation: pathItem.Patch},
			{method: "POST", operation: pathItem.Post},
			{method: "PUT", operation: pathItem.Put},
			{method: "TRACE", operation: pathItem.Trace},
		}
		for _, item := range operations {
			if item.operation == nil {
				continue
			}
			declaration, ok := lookupOpenAPIRouteAuthorization(s.routeAuth, item.method, pathName)
			if !ok || !declaration.AllowsActor(authcap.ActorKindMachine) {
				continue
			}
			applyOpenAPIMachineAuthorization(item.operation, declaration)
		}
	}
}

func lookupOpenAPIRouteAuthorization(
	catalog authcap.RouteAuthorizationCatalogue,
	method string,
	path string,
) (authcap.RouteAuthorization, bool) {
	if catalog == nil {
		return authcap.RouteAuthorization{}, false
	}
	if declaration, ok := catalog.Lookup(method, path); ok {
		return declaration, true
	}

	prefix := pluginhost.PluginAPINamespacePrefix + "/"
	trimmed := strings.TrimPrefix(path, prefix)
	if trimmed == path {
		return authcap.RouteAuthorization{}, false
	}
	separator := strings.Index(trimmed, "/")
	if separator <= 0 {
		return authcap.RouteAuthorization{}, false
	}
	pluginID := trimmed[:separator]
	declaration, ok := catalog.Lookup(method, trimmed[separator:])
	if !ok || declaration.OwnerKind != authcap.RouteOwnerKindDynamicPlugin || declaration.OwnerID != pluginID {
		return authcap.RouteAuthorization{}, false
	}
	return declaration, true
}

func applyOpenAPIMachineAuthorization(operation *goai.Operation, declaration authcap.RouteAuthorization) {
	if operation == nil {
		return
	}
	security := goai.SecurityRequirements{{"LinaHMAC": {}}}
	if declaration.AllowsActor(authcap.ActorKindUser) {
		security = goai.SecurityRequirements{{"BearerAuth": {}}, {"LinaHMAC": {}}}
	}
	operation.Security = &security
	if operation.XExtensions == nil {
		operation.XExtensions = goai.XExtensions{}
	}
	operation.XExtensions[openAPIOperationExtension] = string(declaration.Operation)
	operation.XExtensions[openAPIResourceExtension] = string(declaration.Resource)
	operation.XExtensions[openAPIActionExtension] = string(declaration.Access)
	operation.XExtensions[openAPIActorsExtension] = formatOpenAPIActors(declaration.Actors)
}

func formatOpenAPIActors(actors []authcap.ActorKind) string {
	values := make([]string, 0, len(actors))
	for _, actor := range actors {
		values = append(values, string(actor))
	}
	return strings.Join(values, ",")
}
