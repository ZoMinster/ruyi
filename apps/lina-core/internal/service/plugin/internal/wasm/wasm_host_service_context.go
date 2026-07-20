// This file adapts context-oriented host-service calls to trusted host request
// state and shared business-context capability services.

package wasm

import (
	"context"
	"strings"

	"lina-core/pkg/plugin/capability/bizctxcap"
	"lina-core/pkg/plugin/capability/routecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

// dispatchBizCtxHostService routes business-context host-service calls.
func dispatchBizCtxHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	_ []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	service := bizCtxServiceForHostCall(hcc)
	if service == nil {
		return domainServiceNotScoped("bizctx")
	}
	if method != bridgehostservice.HostServiceMethodBizCtxCurrent {
		return domainMethodNotFound("bizctx", method)
	}
	return capabilityJSONResponse(service.Current(ctx))
}

// bizCtxServiceForHostCall resolves the business context service for one host call.
func bizCtxServiceForHostCall(hcc *hostCallContext) bizctxcap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.BizCtx()
}

// dispatchRouteHostService routes current dynamic-route metadata reads.
func dispatchRouteHostService(
	ctx context.Context,
	hcc *hostCallContext,
	method string,
	payload []byte,
) *bridgehostcall.HostCallResponseEnvelope {
	switch method {
	case bridgehostservice.HostServiceMethodRouteMetadataGet:
		return capabilityJSONResponse(routeMetadataFromHostCall(hcc))
	case bridgehostservice.HostServiceMethodRouteMachineAuthorizationsList:
		service := routeServiceForHostCall(hcc)
		if service == nil {
			return domainServiceNotScoped("route")
		}
		var request routecap.MachineAuthorizationListInput
		if err := decodeCapabilityJSONRequest(payload, &request); err != nil {
			return invalidCapabilityRequest(err)
		}
		result, err := service.ListMachineAuthorizations(ctx, request)
		return domainCapabilityResult(result, err)
	default:
		return domainMethodNotFound("route", method)
	}
}

// routeServiceForHostCall resolves the route capability for one dynamic guest.
func routeServiceForHostCall(hcc *hostCallContext) routecap.Service {
	services := capabilityServicesForHostCall(hcc)
	if services == nil {
		return nil
	}
	return services.Route()
}

// routeMetadataFromHostCall projects trusted host-call context into route metadata.
func routeMetadataFromHostCall(hcc *hostCallContext) *routecap.Metadata {
	if hcc == nil {
		return nil
	}
	return &routecap.Metadata{
		PluginID:   strings.TrimSpace(hcc.pluginID),
		PublicPath: strings.TrimSpace(hcc.routePath),
		Meta: map[string]string{
			"executionSource": strings.TrimSpace(string(hcc.executionSource)),
			"requestId":       strings.TrimSpace(hcc.requestID),
		},
	}
}
