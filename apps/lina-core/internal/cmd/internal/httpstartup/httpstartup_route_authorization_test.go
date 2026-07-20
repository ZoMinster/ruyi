// This file verifies static GoFrame route metadata parsing and machine default denial.

package httpstartup

import (
	"context"
	"net/http"
	"testing"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/plugin/capability/authcap"
)

type staticMachineRouteReq struct {
	g.Meta `path:"/records" method:"get" operation:"host.records.list" resource:"host.records" action:"read" actors:"user,machine"`
}

type staticMachineRouteRes struct{}

type staticUserRouteReq struct {
	g.Meta `path:"/profile" method:"get"`
}

type staticUserRouteRes struct{}

func staticMachineRouteHandler(context.Context, *staticMachineRouteReq) (*staticMachineRouteRes, error) {
	return &staticMachineRouteRes{}, nil
}

func staticUserRouteHandler(context.Context, *staticUserRouteReq) (*staticUserRouteRes, error) {
	return &staticUserRouteRes{}, nil
}

// TestParseHandlerRouteAuthorization verifies static handlers use the same
// normalized metadata semantics as source and dynamic plugin declarations.
func TestParseHandlerRouteAuthorization(t *testing.T) {
	t.Parallel()

	machine, err := parseHandlerRouteAuthorization(
		staticMachineRouteHandler,
		authcap.RouteOwnerKindHost,
		hostRouteAuthorizationOwnerID,
		http.MethodGet,
		"/records",
	)
	if err != nil {
		t.Fatalf("parse static machine route: %v", err)
	}
	if !machine.AllowsActor(authcap.ActorKindMachine) || machine.Operation != "host.records.list" {
		t.Fatalf("unexpected machine route metadata: %#v", machine)
	}

	userOnly, err := parseHandlerRouteAuthorization(
		staticUserRouteHandler,
		authcap.RouteOwnerKindHost,
		hostRouteAuthorizationOwnerID,
		http.MethodGet,
		"/profile",
	)
	if err != nil {
		t.Fatalf("parse static user route: %v", err)
	}
	if userOnly.AllowsActor(authcap.ActorKindMachine) || !userOnly.AllowsActor(authcap.ActorKindUser) {
		t.Fatalf("expected missing actors to default user-only, got %#v", userOnly.Actors)
	}
}
