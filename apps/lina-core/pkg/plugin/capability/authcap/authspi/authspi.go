// Package authspi defines the source-plugin machine-authentication provider
// contract. It is separate from the ordinary authcap.Service consumer surface:
// only host startup and provider plugins use this package to register and invoke
// authentication implementations.
package authspi

import (
	"context"

	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/plugin/capability/authcap/machinecoord"
	"lina-core/pkg/plugin/capability/plugincap"
)

// Provider authenticates requests for one authorization scheme. The host
// selects the provider before calling Authenticate; implementations must never
// fall back to another scheme, create user sessions, or return reusable secret
// material in the result or error.
type Provider interface {
	// Authenticate validates one detached request projection and returns a
	// machine actor plus its bounded authorization snapshot. Authentication
	// failures return an error and no partially trusted result.
	Authenticate(ctx context.Context, request authcap.AuthenticationRequest) (authcap.AuthenticationResult, error)
}

// Dispatcher resolves one exact authorization scheme and delegates to its
// current enabled provider without fallback.
type Dispatcher interface {
	// Authenticate authenticates one detached request through the provider
	// registered for scheme. Unknown and unavailable schemes return an error.
	Authenticate(
		ctx context.Context,
		scheme string,
		request authcap.AuthenticationRequest,
	) (authcap.AuthenticationResult, error)
}

// ProviderEnv carries host-stamped values available while constructing one
// provider. Runtime capability dependencies are added only when the host owns
// and can inject a stable governed contract for them.
type ProviderEnv struct {
	// PluginID is the provider plugin identity stamped by the host registry.
	PluginID string
	// MachineCoordination is bound to PluginID and exposes only machine-access
	// revision and replay coordination operations.
	MachineCoordination machinecoord.Service
	// Config is the provider plugin's scoped runtime configuration reader.
	Config plugincap.ConfigService
}

// ProviderFactory creates one provider from its host-stamped environment.
// Factories run through startup-owned provider management and must return an
// error when required dependencies or configuration are unavailable.
type ProviderFactory func(ctx context.Context, env ProviderEnv) (Provider, error)
