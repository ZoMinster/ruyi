// This file defines the source-plugin visible dynamic-route contract.

package routecap

import "context"

const (
	// DefaultMachineAuthorizationLimit bounds ordinary catalog reads when the
	// caller does not request a stricter limit.
	DefaultMachineAuthorizationLimit = 2048
	// MaxMachineAuthorizationLimit is the hard upper bound for one catalog read.
	MaxMachineAuthorizationLimit = 10000
)

// Service defines dynamic-route metadata operations published to source plugins.
type Service interface {
	// GetMetadata returns metadata attached to the current dynamic-route request.
	GetMetadata(ctx context.Context) *Metadata
	// ListMachineAuthorizations returns one bounded snapshot of all declared
	// machine routes and their resource access modes. Active reflects the current
	// host or tenant plugin enablement state; inactive declarations remain in the
	// snapshot so policy relations can survive plugin disable and re-enable.
	ListMachineAuthorizations(ctx context.Context, input MachineAuthorizationListInput) (*MachineAuthorizationCatalogue, error)
}

// MachineAuthorizationListInput controls one bounded machine-route catalog read.
type MachineAuthorizationListInput struct {
	// Limit is the maximum number of route declarations accepted in one snapshot.
	// Zero uses DefaultMachineAuthorizationLimit.
	Limit int `json:"limit"`
}

// MachineAuthorizationCatalogue is a detached route and resource projection.
type MachineAuthorizationCatalogue struct {
	// Routes contains stable operation-level declarations in deterministic order.
	Routes []MachineRouteAuthorization `json:"routes"`
	// Resources contains stable resource codes and declared/effective access modes.
	Resources []MachineResourceAuthorization `json:"resources"`
	// Total is the number of machine route declarations in this snapshot.
	Total int `json:"total"`
}

// MachineRouteAuthorization is one machine-enabled route declaration.
type MachineRouteAuthorization struct {
	// OwnerKind identifies host, source-plugin, or dynamic-plugin ownership.
	OwnerKind string `json:"ownerKind"`
	// OwnerID identifies the host component or plugin that owns the route.
	OwnerID string `json:"ownerId"`
	// Method is the normalized HTTP method.
	Method string `json:"method"`
	// Path is the normalized registered route pattern.
	Path string `json:"path"`
	// Operation is the globally unique stable machine interface code.
	Operation string `json:"operation"`
	// Resource is the stable whole-resource type code.
	Resource string `json:"resource"`
	// Action is the resource-wide read or write mode.
	Action string `json:"action"`
	// Active reports whether the route owner is currently enabled in scope.
	Active bool `json:"active"`
}

// MachineResourceAuthorization aggregates declared and currently active route
// actions for one stable whole-resource type.
type MachineResourceAuthorization struct {
	// Resource is the stable whole-resource type code.
	Resource string `json:"resource"`
	// Read reports whether any route declares read access for this resource.
	Read bool `json:"read"`
	// Write reports whether any route declares write access for this resource.
	Write bool `json:"write"`
	// ActiveRead reports whether a currently active route reads this resource.
	ActiveRead bool `json:"activeRead"`
	// ActiveWrite reports whether a currently active route writes this resource.
	ActiveWrite bool `json:"activeWrite"`
}

// Metadata is the published metadata of one matched dynamic route.
type Metadata struct {
	// PluginID is the dynamic plugin that owns the matched route.
	PluginID string
	// Method is the declared dynamic route HTTP method.
	Method string
	// PublicPath is the public host path matched by the request.
	PublicPath string
	// Tags are the route tags declared by the dynamic plugin manifest.
	Tags []string
	// Summary is the route summary declared by the dynamic plugin manifest.
	Summary string
	// Operation is the globally unique stable machine interface operation code.
	Operation string
	// Resource is the stable whole-resource type governed by the route.
	Resource string
	// Action is the resource-wide read or write action.
	Action string
	// Actors is the normalized comma-separated user/machine allowlist.
	Actors string
	// Meta contains additional route declaration metadata by source tag name.
	Meta map[string]string
	// ResponseBody stores the raw bridge response body captured by the runtime dispatcher.
	ResponseBody string
	// ResponseContentType stores the resolved content type of the bridge response.
	ResponseContentType string
}
