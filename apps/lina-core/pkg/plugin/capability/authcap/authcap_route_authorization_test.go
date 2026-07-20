// This file verifies route metadata parsing, machine default denial, global
// operation uniqueness, and atomic owner replacement behavior.

package authcap

import "testing"

// TestParseRouteAuthorizationDefaultsToUser verifies missing actors never
// implicitly exposes an existing interface to machine principals.
func TestParseRouteAuthorizationDefaultsToUser(t *testing.T) {
	t.Parallel()

	item, err := ParseRouteAuthorization(RouteOwnerKindHost, "core", "get", "/records", "", "", "", "")
	if err != nil {
		t.Fatalf("parse user-default route: %v", err)
	}
	if !item.AllowsActor(ActorKindUser) || item.AllowsActor(ActorKindMachine) {
		t.Fatalf("expected user-only route, got %#v", item.Actors)
	}
}

// TestParseRouteAuthorizationRejectsInvalidMachineMetadata verifies machine
// declarations fail closed on missing metadata and unknown enum values.
func TestParseRouteAuthorizationRejectsInvalidMachineMetadata(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		resource string
		action   string
		actors   string
	}{
		{name: "missing resource", action: "read", actors: "machine"},
		{name: "unknown action", resource: "records", action: "execute", actors: "machine"},
		{name: "unknown actor", resource: "records", action: "read", actors: "service"},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			_, err := ParseRouteAuthorization(
				RouteOwnerKindHost,
				"core",
				"GET",
				"/records",
				"records.list",
				testCase.resource,
				testCase.action,
				testCase.actors,
			)
			if err == nil {
				t.Fatal("expected invalid machine metadata to fail")
			}
		})
	}
}

// TestRouteAuthorizationCatalogueRejectsDuplicateOperation verifies operation
// codes remain globally unique across route owners.
func TestRouteAuthorizationCatalogueRejectsDuplicateOperation(t *testing.T) {
	t.Parallel()

	catalog := NewRouteAuthorizationCatalogue()
	first := mustMachineRoute(t, RouteOwnerKindSourcePlugin, "plugin-a", "GET", "/a", "records.list")
	second := mustMachineRoute(t, RouteOwnerKindDynamicPlugin, "plugin-b", "GET", "/b", "records.list")
	if err := catalog.ReplaceOwner(first.OwnerKey(), []RouteAuthorization{first}); err != nil {
		t.Fatalf("publish first route: %v", err)
	}
	if err := catalog.ReplaceOwner(second.OwnerKey(), []RouteAuthorization{second}); err == nil {
		t.Fatal("expected duplicate global operation to fail")
	}
	if routes := catalog.ListMachineRoutes(); len(routes) != 1 || routes[0].Path != "/a" {
		t.Fatalf("failed replacement must retain original snapshot, got %#v", routes)
	}
}

// TestRouteAuthorizationCatalogueReplacesOwnerAtomically verifies upgrades
// remove stale owner routes and publish the complete replacement snapshot.
func TestRouteAuthorizationCatalogueReplacesOwnerAtomically(t *testing.T) {
	t.Parallel()

	catalog := NewRouteAuthorizationCatalogue()
	first := mustMachineRoute(t, RouteOwnerKindSourcePlugin, "plugin-a", "GET", "/v1/records", "records.list.v1")
	second := mustMachineRoute(t, RouteOwnerKindSourcePlugin, "plugin-a", "GET", "/v2/records", "records.list.v2")
	if err := catalog.ReplaceOwner(first.OwnerKey(), []RouteAuthorization{first}); err != nil {
		t.Fatalf("publish first owner snapshot: %v", err)
	}
	if err := catalog.ReplaceOwner(second.OwnerKey(), []RouteAuthorization{second}); err != nil {
		t.Fatalf("replace owner snapshot: %v", err)
	}
	if _, ok := catalog.Lookup("GET", "/v1/records"); ok {
		t.Fatal("expected stale owner route to be removed")
	}
	if item, ok := catalog.Lookup("get", "/v2/records/"); !ok || item.Operation != second.Operation {
		t.Fatalf("expected replacement route lookup, got %#v ok=%v", item, ok)
	}
}

func mustMachineRoute(
	t *testing.T,
	kind RouteOwnerKind,
	owner string,
	method string,
	path string,
	operation string,
) RouteAuthorization {
	t.Helper()
	item, err := ParseRouteAuthorization(
		kind,
		owner,
		method,
		path,
		operation,
		"records",
		"read",
		"user,machine",
	)
	if err != nil {
		t.Fatalf("build machine route: %v", err)
	}
	return item
}
