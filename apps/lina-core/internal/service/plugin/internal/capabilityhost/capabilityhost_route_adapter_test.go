// This file verifies bounded machine-route projections, batch owner-state
// resolution, and disable/re-enable behavior for source and dynamic plugins.

package capabilityhost

import (
	"context"
	"reflect"
	"testing"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/routecap"
)

type routeCataloguePluginState struct {
	enabled   map[string]bool
	requested []string
}

func (s *routeCataloguePluginState) IsEnabled(_ context.Context, pluginID string) bool {
	return s.enabled[pluginID]
}

func (s *routeCataloguePluginState) ResolveBusinessEntryEnablement(
	_ context.Context,
	pluginIDs []string,
) (map[string]bool, error) {
	s.requested = append([]string(nil), pluginIDs...)
	result := make(map[string]bool, len(pluginIDs))
	for _, pluginID := range pluginIDs {
		result[pluginID] = s.enabled[pluginID]
	}
	return result, nil
}

func (s *routeCataloguePluginState) IsProviderEnabled(context.Context, string) bool { return false }

func (s *routeCataloguePluginState) IsEnabledAuthoritative(context.Context, string) bool {
	return false
}

func TestRouteAdapterListsBoundedMachineAuthorizationsAndRestoresEnablement(t *testing.T) {
	catalog := authcap.NewRouteAuthorizationCatalogue()
	routes := []authcap.RouteAuthorization{
		mustMachineRoute(t, authcap.RouteOwnerKindHost, "lina-core", "GET", "/core/resources", "core.resources.list", "core-resources", "read"),
		mustMachineRoute(t, authcap.RouteOwnerKindSourcePlugin, "plugin-orders", "GET", "/orders", "orders.list", "orders", "read"),
		mustMachineRoute(t, authcap.RouteOwnerKindDynamicPlugin, "plugin-orders-worker", "POST", "/orders", "orders.create", "orders", "write"),
	}
	for _, route := range routes {
		if err := catalog.ReplaceOwner(route.OwnerKey(), routesForOwner(routes, route.OwnerKey())); err != nil {
			t.Fatalf("publish route owner %s: %v", route.OwnerKey(), err)
		}
	}
	state := &routeCataloguePluginState{enabled: map[string]bool{
		"plugin-orders":        true,
		"plugin-orders-worker": false,
	}}
	service := newRouteAdapter(catalog, state)

	output, err := service.ListMachineAuthorizations(context.Background(), routecap.MachineAuthorizationListInput{Limit: 3})
	if err != nil {
		t.Fatalf("list machine authorizations: %v", err)
	}
	if output.Total != 3 || len(output.Routes) != 3 || len(output.Resources) != 2 {
		t.Fatalf("unexpected catalog projection: %#v", output)
	}
	if !reflect.DeepEqual(state.requested, []string{"plugin-orders", "plugin-orders-worker"}) {
		t.Fatalf("expected one stable batch of plugin owners, got %#v", state.requested)
	}
	if !routeByOperation(t, output, "core.resources.list").Active || !routeByOperation(t, output, "orders.list").Active {
		t.Fatal("expected host and enabled source routes to be active")
	}
	if routeByOperation(t, output, "orders.create").Active {
		t.Fatal("expected disabled dynamic route declaration to remain but be inactive")
	}
	orders := resourceByCode(t, output, "orders")
	if !orders.Read || !orders.Write || !orders.ActiveRead || orders.ActiveWrite {
		t.Fatalf("unexpected declared/effective orders modes: %#v", orders)
	}
	if _, err = service.ListMachineAuthorizations(context.Background(), routecap.MachineAuthorizationListInput{Limit: 2}); err == nil {
		t.Fatal("expected catalog limit overflow to fail closed")
	}

	state.enabled["plugin-orders-worker"] = true
	reenabled, err := service.ListMachineAuthorizations(context.Background(), routecap.MachineAuthorizationListInput{Limit: 3})
	if err != nil {
		t.Fatalf("list re-enabled machine authorizations: %v", err)
	}
	if !routeByOperation(t, reenabled, "orders.create").Active || !resourceByCode(t, reenabled, "orders").ActiveWrite {
		t.Fatal("expected existing declarations to recover after plugin re-enable")
	}
}

func mustMachineRoute(
	t *testing.T,
	ownerKind authcap.RouteOwnerKind,
	ownerID string,
	method string,
	path string,
	operation string,
	resource string,
	action string,
) authcap.RouteAuthorization {
	t.Helper()
	route, err := authcap.ParseRouteAuthorization(ownerKind, ownerID, method, path, operation, resource, action, "machine")
	if err != nil {
		t.Fatalf("parse machine route %s: %v", operation, err)
	}
	return route
}

func routesForOwner(routes []authcap.RouteAuthorization, ownerKey string) []authcap.RouteAuthorization {
	items := make([]authcap.RouteAuthorization, 0)
	for _, route := range routes {
		if route.OwnerKey() == ownerKey {
			items = append(items, route)
		}
	}
	return items
}

func routeByOperation(
	t *testing.T,
	catalog *routecap.MachineAuthorizationCatalogue,
	operation string,
) routecap.MachineRouteAuthorization {
	t.Helper()
	for _, route := range catalog.Routes {
		if route.Operation == operation {
			return route
		}
	}
	t.Fatalf("route operation %s not found", operation)
	return routecap.MachineRouteAuthorization{}
}

func resourceByCode(
	t *testing.T,
	catalog *routecap.MachineAuthorizationCatalogue,
	resourceCode string,
) routecap.MachineResourceAuthorization {
	t.Helper()
	for _, resource := range catalog.Resources {
		if resource.Resource == resourceCode {
			return resource
		}
	}
	t.Fatalf("resource %s not found", resourceCode)
	return routecap.MachineResourceAuthorization{}
}
