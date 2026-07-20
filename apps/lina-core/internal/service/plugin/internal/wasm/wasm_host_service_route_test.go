// This file verifies dynamic plugins receive the same bounded machine-route
// catalog contract as source plugins through the governed host:route method.

package wasm

import (
	"context"
	"encoding/json"
	"testing"

	"lina-core/pkg/plugin/capability/routecap"
	bridgehostcall "lina-core/pkg/plugin/pluginbridge/protocol"
	bridgehostservice "lina-core/pkg/plugin/pluginbridge/protocol"
)

type routeCatalogueHostService struct {
	input  routecap.MachineAuthorizationListInput
	output *routecap.MachineAuthorizationCatalogue
}

func (s *routeCatalogueHostService) GetMetadata(context.Context) *routecap.Metadata { return nil }

func (s *routeCatalogueHostService) ListMachineAuthorizations(
	_ context.Context,
	input routecap.MachineAuthorizationListInput,
) (*routecap.MachineAuthorizationCatalogue, error) {
	s.input = input
	return s.output, nil
}

func TestDispatchRouteMachineAuthorizationsUsesScopedCapability(t *testing.T) {
	service := &routeCatalogueHostService{output: &routecap.MachineAuthorizationCatalogue{
		Routes: []routecap.MachineRouteAuthorization{{
			OwnerKind: "dynamic", OwnerID: "plugin-orders", Method: "GET", Path: "/orders",
			Operation: "orders.list", Resource: "orders", Action: "read", Active: true,
		}},
		Resources: []routecap.MachineResourceAuthorization{{
			Resource: "orders", Read: true, ActiveRead: true,
		}},
		Total: 1,
	}}
	configureDomainHostServicesForCapabilityTest(t, &capabilityHostServiceTestServices{route: service})
	hcc := withTestHostCallRuntime(t, &hostCallContext{pluginID: "catalog-consumer"})
	payload, err := json.Marshal(routecap.MachineAuthorizationListInput{Limit: 64})
	if err != nil {
		t.Fatalf("marshal route catalog request: %v", err)
	}
	response := dispatchRouteHostService(
		context.Background(),
		hcc,
		bridgehostservice.HostServiceMethodRouteMachineAuthorizationsList,
		bridgehostservice.MarshalHostServiceJSONRequest(&bridgehostservice.HostServiceJSONRequest{Value: payload}),
	)
	if response == nil || response.Status != bridgehostcall.HostCallStatusSuccess {
		t.Fatalf("expected route catalog success, got %#v", response)
	}
	var output routecap.MachineAuthorizationCatalogue
	decodeCapabilityJSONResponse(t, response.Payload, &output)
	if service.input.Limit != 64 || output.Total != 1 || len(output.Routes) != 1 || output.Routes[0].Operation != "orders.list" {
		t.Fatalf("unexpected dynamic route catalog projection: input=%#v output=%#v", service.input, output)
	}
}
