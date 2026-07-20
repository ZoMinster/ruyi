// This file defines neutral request, result, and authorization projections for
// authentication providers. It uses only standard-library-compatible values
// and never exposes GoFrame requests or host-internal models.

package authcap

import "strings"

// AccessMode identifies the resource-wide access class declared by a route.
type AccessMode string

const (
	// AccessModeRead identifies read-only resource operations.
	AccessModeRead AccessMode = "read"
	// AccessModeWrite identifies resource-mutating or executing operations.
	AccessModeWrite AccessMode = "write"
)

// OperationCode identifies one globally unique machine-accessible interface.
type OperationCode string

// ResourceCode identifies one stable resource type governed as a whole.
type ResourceCode string

// QueryParameter preserves one decoded query key/value pair. Repeated keys are
// represented by repeated entries so providers can apply their canonical sort.
type QueryParameter struct {
	// Key is the decoded query parameter name.
	Key string
	// Value is one decoded value for Key.
	Value string
}

// Header preserves one request header and all of its values.
type Header struct {
	// Name is the canonical or original header name.
	Name string
	// Values contains the header values in transport order.
	Values []string
}

// AuthenticationRequest is a detached, read-only request projection supplied
// to one precisely selected authentication provider. Callers construct it with
// NewAuthenticationRequest; getters always return detached slices.
type AuthenticationRequest struct {
	scheme        string
	credential    string
	method        string
	escapedPath   string
	query         []QueryParameter
	headers       []Header
	bodySHA256    string
	remoteAddress string
}

// NewAuthenticationRequest creates a detached provider request projection.
// credential is the scheme payload after the authorization scheme token and
// is sensitive request-scoped data that must not be logged or persisted.
func NewAuthenticationRequest(
	scheme string,
	credential string,
	method string,
	escapedPath string,
	query []QueryParameter,
	headers []Header,
	bodySHA256 string,
	remoteAddress string,
) AuthenticationRequest {
	return AuthenticationRequest{
		scheme:        strings.TrimSpace(scheme),
		credential:    credential,
		method:        strings.ToUpper(strings.TrimSpace(method)),
		escapedPath:   escapedPath,
		query:         cloneQueryParameters(query),
		headers:       cloneHeaders(headers),
		bodySHA256:    strings.ToLower(strings.TrimSpace(bodySHA256)),
		remoteAddress: remoteAddress,
	}
}

// Scheme returns the normalized authorization scheme selected by the host.
func (r AuthenticationRequest) Scheme() string { return r.scheme }

// Credential returns the sensitive scheme payload for provider parsing.
func (r AuthenticationRequest) Credential() string { return r.credential }

// Method returns the normalized uppercase HTTP method.
func (r AuthenticationRequest) Method() string { return r.method }

// EscapedPath returns the transport-visible escaped request path.
func (r AuthenticationRequest) EscapedPath() string { return r.escapedPath }

// Query returns a detached query parameter list preserving repeated keys.
func (r AuthenticationRequest) Query() []QueryParameter { return cloneQueryParameters(r.query) }

// Headers returns detached request headers selected by the host projection.
func (r AuthenticationRequest) Headers() []Header { return cloneHeaders(r.headers) }

// Header returns the first case-insensitive value for name.
func (r AuthenticationRequest) Header(name string) string {
	for _, header := range r.headers {
		if strings.EqualFold(strings.TrimSpace(header.Name), strings.TrimSpace(name)) && len(header.Values) > 0 {
			return header.Values[0]
		}
	}
	return ""
}

// BodySHA256 returns the host-computed lowercase hexadecimal request-body hash.
func (r AuthenticationRequest) BodySHA256() string { return r.bodySHA256 }

// RemoteAddress returns the transport peer address captured by the host.
func (r AuthenticationRequest) RemoteAddress() string { return r.remoteAddress }

// ResourcePermission is one resource-wide read or write grant.
type ResourcePermission struct {
	// Resource identifies the governed resource type.
	Resource ResourceCode
	// Access identifies the granted resource-wide access mode.
	Access AccessMode
}

// AuthorizationRequest describes the exact route and resource access that a
// machine actor must be allowed to perform.
type AuthorizationRequest struct {
	// Operation identifies the exact interface operation.
	Operation OperationCode
	// Resource identifies the whole resource type operated on by the route.
	Resource ResourceCode
	// Access identifies whether the route reads or writes the resource type.
	Access AccessMode
}

// AuthorizationSnapshot is an immutable-by-contract allow-only projection.
// It stores bounded operation and resource sets and applies default-deny,
// requiring both grants for every machine authorization decision.
type AuthorizationSnapshot struct {
	operations map[OperationCode]struct{}
	resources  map[ResourceCode]map[AccessMode]struct{}
}

// NewAuthorizationSnapshot builds a detached allow-only authorization snapshot.
func NewAuthorizationSnapshot(operations []OperationCode, resources []ResourcePermission) AuthorizationSnapshot {
	snapshot := AuthorizationSnapshot{
		operations: make(map[OperationCode]struct{}, len(operations)),
		resources:  make(map[ResourceCode]map[AccessMode]struct{}),
	}
	for _, operation := range operations {
		if operation != "" {
			snapshot.operations[operation] = struct{}{}
		}
	}
	for _, permission := range resources {
		if permission.Resource == "" || (permission.Access != AccessModeRead && permission.Access != AccessModeWrite) {
			continue
		}
		if snapshot.resources[permission.Resource] == nil {
			snapshot.resources[permission.Resource] = make(map[AccessMode]struct{}, 2)
		}
		snapshot.resources[permission.Resource][permission.Access] = struct{}{}
	}
	return snapshot
}

// Allows reports whether both the exact interface and resource-wide access
// grants are present. Empty or unknown values always fail closed.
func (s AuthorizationSnapshot) Allows(request AuthorizationRequest) bool {
	if request.Operation == "" || request.Resource == "" {
		return false
	}
	if request.Access != AccessModeRead && request.Access != AccessModeWrite {
		return false
	}
	if _, ok := s.operations[request.Operation]; !ok {
		return false
	}
	_, ok := s.resources[request.Resource][request.Access]
	return ok
}

// AuthenticationResult contains the trusted actor and bounded allow-only
// authorization snapshot returned by one machine authentication provider.
type AuthenticationResult struct {
	// Actor is the principal established by the selected provider.
	Actor Actor
	// Authorization contains the exact interface and resource-wide grants used
	// by host authorization middleware.
	Authorization AuthorizationSnapshot
}

// cloneQueryParameters detaches a query projection.
func cloneQueryParameters(values []QueryParameter) []QueryParameter {
	if len(values) == 0 {
		return nil
	}
	return append([]QueryParameter(nil), values...)
}

// cloneHeaders detaches header names and nested value slices.
func cloneHeaders(values []Header) []Header {
	if len(values) == 0 {
		return nil
	}
	cloned := make([]Header, len(values))
	for i, header := range values {
		cloned[i] = Header{Name: header.Name, Values: append([]string(nil), header.Values...)}
	}
	return cloned
}
