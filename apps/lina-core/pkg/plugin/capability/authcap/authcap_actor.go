// This file defines authentication actor values shared by host middleware,
// source plugins, and dynamic-plugin projections without exposing user-session
// models or machine-provider internals.

package authcap

// ActorKind identifies the trusted principal category established by the host
// authentication chain.
type ActorKind string

const (
	// ActorKindUser identifies an interactive user authenticated by the host.
	ActorKindUser ActorKind = "user"
	// ActorKindMachine identifies a non-user machine client authenticated by a
	// registered machine authentication provider.
	ActorKindMachine ActorKind = "machine"
)

// Actor is the host-trusted principal projection used by authorization and
// plugin capability contexts. SubjectID and CredentialID are opaque stable
// identifiers owned by the selected authentication provider. The projection
// intentionally has no user ID, username, role, or session fields, so a
// machine actor cannot be represented as a user session.
type Actor struct {
	// Kind identifies whether the principal is a user or a machine.
	Kind ActorKind
	// SubjectID identifies the authenticated user or machine client.
	SubjectID string
	// CredentialID identifies the token, session, or machine credential used for
	// this request without containing reusable credential material.
	CredentialID string
	// TenantID is the trusted tenant boundary established by authentication;
	// zero identifies platform scope.
	TenantID int
}
