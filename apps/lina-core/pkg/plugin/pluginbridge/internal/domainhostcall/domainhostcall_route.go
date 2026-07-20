// This file implements the guest-side route capability hostcall client.
// The host remains authoritative for dynamic route metadata fields.

package domainhostcall

import (
	"context"

	"lina-core/pkg/plugin/capability/routecap"
	"lina-core/pkg/plugin/pluginbridge/protocol"
)

// routeService adapts dynamic route metadata reads to host services.
type routeService struct{ baseService }

// Route creates the current dynamic-route metadata guest client.
func Route(invoker Invoker) routecap.Service {
	return routeService{baseService: newBaseService(invoker)}
}

// GetMetadata returns the current dynamic-route metadata projection.
func (s routeService) GetMetadata(context.Context) *routecap.Metadata {
	var out routecap.Metadata
	if err := s.callJSONRequest(protocol.HostServiceRoute, protocol.HostServiceMethodRouteMetadataGet, nil, &out); err != nil {
		return nil
	}
	return &out
}

// ListMachineAuthorizations returns the host-owned bounded machine route catalog.
func (s routeService) ListMachineAuthorizations(
	_ context.Context,
	input routecap.MachineAuthorizationListInput,
) (*routecap.MachineAuthorizationCatalogue, error) {
	var out routecap.MachineAuthorizationCatalogue
	err := s.callJSONRequest(
		protocol.HostServiceRoute,
		protocol.HostServiceMethodRouteMachineAuthorizationsList,
		input,
		&out,
	)
	return &out, err
}

var _ routecap.Service = (*routeService)(nil)
