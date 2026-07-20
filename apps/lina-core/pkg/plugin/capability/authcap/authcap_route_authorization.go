// This file defines route-level machine authorization metadata and the
// startup-owned atomic catalog shared by host and plugin route lifecycles.

package authcap

import (
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"
)

// RouteOwnerKind identifies the route publication mechanism that owns a
// catalog entry.
type RouteOwnerKind string

const (
	// RouteOwnerKindHost identifies host-owned static API routes.
	RouteOwnerKindHost RouteOwnerKind = "host"
	// RouteOwnerKindSourcePlugin identifies source-plugin routes.
	RouteOwnerKindSourcePlugin RouteOwnerKind = "source"
	// RouteOwnerKindDynamicPlugin identifies dynamic-plugin routes.
	RouteOwnerKindDynamicPlugin RouteOwnerKind = "dynamic"
)

// RouteAuthorization describes the stable authorization identity of one HTTP
// route. Actors defaults to user-only when parsed from an empty declaration.
type RouteAuthorization struct {
	// OwnerKind identifies the route publication mechanism.
	OwnerKind RouteOwnerKind
	// OwnerID identifies the host or plugin that owns the route.
	OwnerID string
	// Method is the normalized HTTP method.
	Method string
	// Path is the registered route pattern.
	Path string
	// Operation is the globally unique machine interface operation code.
	Operation OperationCode
	// Resource is the stable whole-resource type code.
	Resource ResourceCode
	// Access is the resource-wide read or write action.
	Access AccessMode
	// Actors is the explicit normalized set of principals allowed by the route.
	Actors []ActorKind
}

// OwnerKey returns the stable owner-scoped replacement key.
func (r RouteAuthorization) OwnerKey() string {
	return RouteAuthorizationOwnerKey(r.OwnerKind, r.OwnerID)
}

// RouteKey returns the normalized method and route-pattern lookup key.
func (r RouteAuthorization) RouteKey() string {
	return normalizeRouteMethod(r.Method) + " " + normalizeRoutePath(r.Path)
}

// AllowsActor reports whether the route explicitly allows the supplied actor.
func (r RouteAuthorization) AllowsActor(actor ActorKind) bool {
	for _, allowed := range r.Actors {
		if allowed == actor {
			return true
		}
	}
	return false
}

// RouteAuthorizationOwnerKey returns one stable owner-scoped replacement key.
func RouteAuthorizationOwnerKey(kind RouteOwnerKind, ownerID string) string {
	return string(kind) + ":" + strings.TrimSpace(ownerID)
}

// ParseRouteAuthorization parses raw route tags into a normalized declaration.
// Missing actors intentionally defaults to user-only. Any route that allows a
// machine must declare a complete operation, resource, and read/write action.
func ParseRouteAuthorization(
	ownerKind RouteOwnerKind,
	ownerID string,
	method string,
	path string,
	operation string,
	resource string,
	action string,
	actors string,
) (RouteAuthorization, error) {
	declaration := RouteAuthorization{
		OwnerKind: ownerKind,
		OwnerID:   strings.TrimSpace(ownerID),
		Method:    normalizeRouteMethod(method),
		Path:      normalizeRoutePath(path),
		Operation: OperationCode(strings.TrimSpace(operation)),
		Resource:  ResourceCode(strings.TrimSpace(resource)),
		Access:    AccessMode(strings.ToLower(strings.TrimSpace(action))),
	}

	parsedActors, err := parseRouteActors(actors)
	if err != nil {
		return RouteAuthorization{}, err
	}
	declaration.Actors = parsedActors

	if declaration.Access != "" && declaration.Access != AccessModeRead && declaration.Access != AccessModeWrite {
		return RouteAuthorization{}, gerror.Newf(
			"route authorization action only supports read/write: %s %s",
			declaration.Method,
			declaration.Path,
		)
	}
	if declaration.AllowsActor(ActorKindMachine) {
		if declaration.Operation == "" || declaration.Resource == "" || declaration.Access == "" {
			return RouteAuthorization{}, gerror.Newf(
				"machine route requires operation, resource, and action metadata: %s %s",
				declaration.Method,
				declaration.Path,
			)
		}
	}
	return declaration, nil
}

// RouteAuthorizationCatalogue is the shared immutable-snapshot route
// authorization directory. Replacements validate the complete candidate
// snapshot before one atomic publication, so readers never observe partial
// owner upgrades or duplicate operation codes.
type RouteAuthorizationCatalogue interface {
	// ValidateOwner validates one owner replacement against the current global
	// snapshot without publishing it.
	ValidateOwner(ownerKey string, routes []RouteAuthorization) error
	// ReplaceOwner atomically replaces all routes published by one owner.
	ReplaceOwner(ownerKey string, routes []RouteAuthorization) error
	// ReplaceAll atomically replaces the complete owner-to-routes snapshot.
	ReplaceAll(routesByOwner map[string][]RouteAuthorization) error
	// RemoveOwner atomically removes all routes published by one owner.
	RemoveOwner(ownerKey string)
	// Lookup resolves one route by normalized method and route pattern.
	Lookup(method string, path string) (RouteAuthorization, bool)
	// ListMachineRoutes returns a detached, stable-order machine route snapshot.
	ListMachineRoutes() []RouteAuthorization
}

type routeAuthorizationSnapshot struct {
	byOwner     map[string][]RouteAuthorization
	byRoute     map[string]RouteAuthorization
	machineList []RouteAuthorization
}

// routeAuthorizationCatalogue serializes writers while serving lock-free
// immutable snapshots to request and management readers.
type routeAuthorizationCatalogue struct {
	writeMu  sync.Mutex
	snapshot atomic.Pointer[routeAuthorizationSnapshot]
}

// Ensure routeAuthorizationCatalogue implements the public contract.
var _ RouteAuthorizationCatalogue = (*routeAuthorizationCatalogue)(nil)

// NewRouteAuthorizationCatalogue creates an empty atomic route catalog.
func NewRouteAuthorizationCatalogue() RouteAuthorizationCatalogue {
	catalog := &routeAuthorizationCatalogue{}
	catalog.snapshot.Store(emptyRouteAuthorizationSnapshot())
	return catalog
}

// ReplaceOwner atomically replaces all routes published by one owner.
func (c *routeAuthorizationCatalogue) ReplaceOwner(ownerKey string, routes []RouteAuthorization) error {
	trimmedOwner := strings.TrimSpace(ownerKey)
	if trimmedOwner == "" {
		return gerror.New("route authorization owner key cannot be empty")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	next := cloneRouteAuthorizationOwners(c.load().byOwner)
	next[trimmedOwner] = cloneRouteAuthorizations(routes)
	snapshot, err := buildRouteAuthorizationSnapshot(next)
	if err != nil {
		return err
	}
	c.snapshot.Store(snapshot)
	return nil
}

// ValidateOwner validates one owner replacement against the current global
// snapshot without publishing it.
func (c *routeAuthorizationCatalogue) ValidateOwner(ownerKey string, routes []RouteAuthorization) error {
	trimmedOwner := strings.TrimSpace(ownerKey)
	if trimmedOwner == "" {
		return gerror.New("route authorization owner key cannot be empty")
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	next := cloneRouteAuthorizationOwners(c.load().byOwner)
	next[trimmedOwner] = cloneRouteAuthorizations(routes)
	_, err := buildRouteAuthorizationSnapshot(next)
	return err
}

// ReplaceAll atomically replaces the complete owner-to-routes snapshot.
func (c *routeAuthorizationCatalogue) ReplaceAll(routesByOwner map[string][]RouteAuthorization) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	snapshot, err := buildRouteAuthorizationSnapshot(routesByOwner)
	if err != nil {
		return err
	}
	c.snapshot.Store(snapshot)
	return nil
}

// RemoveOwner atomically removes all routes published by one owner.
func (c *routeAuthorizationCatalogue) RemoveOwner(ownerKey string) {
	trimmedOwner := strings.TrimSpace(ownerKey)
	if trimmedOwner == "" {
		return
	}
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	next := cloneRouteAuthorizationOwners(c.load().byOwner)
	delete(next, trimmedOwner)
	snapshot, err := buildRouteAuthorizationSnapshot(next)
	if err == nil {
		c.snapshot.Store(snapshot)
	}
}

// Lookup resolves one route by normalized method and route pattern.
func (c *routeAuthorizationCatalogue) Lookup(method string, path string) (RouteAuthorization, bool) {
	item, ok := c.load().byRoute[normalizeRouteMethod(method)+" "+normalizeRoutePath(path)]
	if !ok {
		return RouteAuthorization{}, false
	}
	return cloneRouteAuthorization(item), true
}

// ListMachineRoutes returns a detached, stable-order machine route snapshot.
func (c *routeAuthorizationCatalogue) ListMachineRoutes() []RouteAuthorization {
	return cloneRouteAuthorizations(c.load().machineList)
}

func (c *routeAuthorizationCatalogue) load() *routeAuthorizationSnapshot {
	if c == nil {
		return emptyRouteAuthorizationSnapshot()
	}
	if snapshot := c.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return emptyRouteAuthorizationSnapshot()
}

func buildRouteAuthorizationSnapshot(
	routesByOwner map[string][]RouteAuthorization,
) (*routeAuthorizationSnapshot, error) {
	snapshot := emptyRouteAuthorizationSnapshot()
	operations := make(map[OperationCode]RouteAuthorization)

	for ownerKey, routes := range routesByOwner {
		trimmedOwner := strings.TrimSpace(ownerKey)
		if trimmedOwner == "" {
			return nil, gerror.New("route authorization owner key cannot be empty")
		}
		for _, route := range routes {
			item := cloneRouteAuthorization(route)
			if item.OwnerKey() != trimmedOwner {
				return nil, gerror.Newf(
					"route authorization owner mismatch: key=%s route=%s",
					trimmedOwner,
					item.OwnerKey(),
				)
			}
			if !item.AllowsActor(ActorKindMachine) {
				continue
			}
			if item.Operation == "" || item.Resource == "" || (item.Access != AccessModeRead && item.Access != AccessModeWrite) {
				return nil, gerror.Newf("machine route authorization metadata is incomplete: %s", item.RouteKey())
			}
			if previous, exists := operations[item.Operation]; exists {
				return nil, gerror.Newf(
					"machine route operation must be globally unique: %s used by %s and %s",
					item.Operation,
					previous.RouteKey(),
					item.RouteKey(),
				)
			}
			if previous, exists := snapshot.byRoute[item.RouteKey()]; exists {
				return nil, gerror.Newf(
					"machine route method and path cannot be duplicated: %s owned by %s and %s",
					item.RouteKey(),
					previous.OwnerKey(),
					item.OwnerKey(),
				)
			}
			operations[item.Operation] = item
			snapshot.byRoute[item.RouteKey()] = item
			snapshot.machineList = append(snapshot.machineList, item)
		}
		snapshot.byOwner[trimmedOwner] = cloneRouteAuthorizations(routes)
	}
	sort.Slice(snapshot.machineList, func(i, j int) bool {
		return snapshot.machineList[i].Operation < snapshot.machineList[j].Operation
	})
	return snapshot, nil
}

func emptyRouteAuthorizationSnapshot() *routeAuthorizationSnapshot {
	return &routeAuthorizationSnapshot{
		byOwner: make(map[string][]RouteAuthorization),
		byRoute: make(map[string]RouteAuthorization),
	}
}

func parseRouteActors(value string) ([]ActorKind, error) {
	if strings.TrimSpace(value) == "" {
		return []ActorKind{ActorKindUser}, nil
	}
	seen := make(map[ActorKind]struct{}, 2)
	actors := make([]ActorKind, 0, 2)
	for _, part := range strings.Split(value, ",") {
		actor := ActorKind(strings.ToLower(strings.TrimSpace(part)))
		switch actor {
		case ActorKindUser, ActorKindMachine:
		default:
			return nil, gerror.Newf("route authorization actor is unsupported: %s", part)
		}
		if _, exists := seen[actor]; exists {
			continue
		}
		seen[actor] = struct{}{}
		actors = append(actors, actor)
	}
	if len(actors) == 0 {
		return []ActorKind{ActorKindUser}, nil
	}
	return actors, nil
}

func normalizeRouteMethod(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}

func normalizeRoutePath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "/" {
		return "/"
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	return strings.TrimRight(trimmed, "/")
}

func cloneRouteAuthorization(value RouteAuthorization) RouteAuthorization {
	value.Actors = append([]ActorKind(nil), value.Actors...)
	return value
}

func cloneRouteAuthorizations(values []RouteAuthorization) []RouteAuthorization {
	if len(values) == 0 {
		return []RouteAuthorization{}
	}
	cloned := make([]RouteAuthorization, len(values))
	for index, value := range values {
		cloned[index] = cloneRouteAuthorization(value)
	}
	return cloned
}

func cloneRouteAuthorizationOwners(
	values map[string][]RouteAuthorization,
) map[string][]RouteAuthorization {
	cloned := make(map[string][]RouteAuthorization, len(values))
	for owner, routes := range values {
		cloned[owner] = cloneRouteAuthorizations(routes)
	}
	return cloned
}
