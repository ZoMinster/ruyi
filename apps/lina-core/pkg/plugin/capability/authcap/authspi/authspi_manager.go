// This file owns the startup-scoped authentication provider directory. It
// publishes immutable scheme snapshots, gates every provider call by current
// plugin enablement, and supports atomic same-owner factory replacement during
// source-plugin runtime upgrades.

package authspi

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap"
)

// ErrProviderUnavailable reports that a known authentication scheme currently
// has no enabled, usable provider. Callers must fail authentication closed.
var ErrProviderUnavailable = gerror.New("authentication provider is unavailable")

// EnablementReader reports whether a provider plugin is currently allowed to
// serve authentication calls.
type EnablementReader interface {
	// IsProviderEnabled reports current platform provider enablement.
	IsProviderEnabled(ctx context.Context, pluginID string) bool
}

// ProviderEnvFactory creates a host-stamped provider environment for one plugin.
type ProviderEnvFactory func(ctx context.Context, pluginID string) ProviderEnv

// Manager owns authentication scheme registrations and publishes atomic
// immutable lookup snapshots for request dispatch.
type Manager struct {
	mu         sync.Mutex
	enablement EnablementReader
	envFactory ProviderEnvFactory
	snapshot   atomic.Pointer[providerSnapshot]
}

// Ensure Manager implements the request dispatcher contract.
var _ Dispatcher = (*Manager)(nil)

// providerSnapshot is immutable after publication.
type providerSnapshot struct {
	providers map[string]Provider
}

// managedProvider lazily constructs one plugin provider and checks enablement
// before and after every authentication call.
type managedProvider struct {
	mu         sync.Mutex
	pluginID   string
	factory    ProviderFactory
	enablement EnablementReader
	envFactory ProviderEnvFactory
	provider   Provider
}

// Ensure managedProvider implements Provider.
var _ Provider = (*managedProvider)(nil)

// NewManager creates an authentication provider manager from startup-owned
// enablement and environment sources. Both dependencies are required because
// silently accepting a node-local or ungoverned provider would weaken disable
// and upgrade guarantees.
func NewManager(enablement EnablementReader, envFactory ProviderEnvFactory) (*Manager, error) {
	if enablement == nil {
		return nil, gerror.New("authentication provider manager requires enablement reader")
	}
	if envFactory == nil {
		return nil, gerror.New("authentication provider manager requires environment factory")
	}
	manager := &Manager{
		enablement: enablement,
		envFactory: envFactory,
	}
	manager.snapshot.Store(&providerSnapshot{providers: map[string]Provider{}})
	return manager, nil
}

// NormalizeScheme validates and normalizes one case-insensitive HTTP
// authorization scheme token. Only RFC token-compatible letters, digits, and
// hyphens are accepted by this contract.
func NormalizeScheme(scheme string) (string, error) {
	normalized := strings.ToUpper(strings.TrimSpace(scheme))
	if normalized == "" {
		return "", gerror.New("authentication scheme is required")
	}
	for _, current := range []byte(normalized) {
		if current >= 'A' && current <= 'Z' || current >= '0' && current <= '9' || current == '-' {
			continue
		}
		return "", gerror.Newf("authentication scheme contains unsupported byte: %q", current)
	}
	return normalized, nil
}

// RegisterFactory registers one globally unique authentication scheme. The
// published snapshot changes atomically only after validation succeeds.
func (m *Manager) RegisterFactory(pluginID string, scheme string, factory ProviderFactory) error {
	if m == nil {
		return gerror.New("authentication provider manager is nil")
	}
	pluginID = strings.TrimSpace(pluginID)
	if pluginID == "" {
		return gerror.New("authentication provider plugin id is required")
	}
	normalized, err := NormalizeScheme(scheme)
	if err != nil {
		return err
	}
	if factory == nil {
		return gerror.Newf("authentication provider factory is nil: scheme=%s plugin=%s", normalized, pluginID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.loadSnapshot()
	if _, exists := current.providers[normalized]; exists {
		return gerror.Newf("authentication scheme already registered: scheme=%s plugin=%s", normalized, pluginID)
	}
	providers := cloneProviders(current.providers)
	providers[normalized] = &managedProvider{
		pluginID:   pluginID,
		factory:    factory,
		enablement: m.enablement,
		envFactory: m.envFactory,
	}
	m.snapshot.Store(&providerSnapshot{providers: providers})
	return nil
}

// ReplaceFactory atomically replaces one scheme factory during a same-owner
// source-plugin runtime upgrade. A different plugin cannot take over an
// existing scheme through replacement.
func (m *Manager) ReplaceFactory(pluginID string, scheme string, factory ProviderFactory) error {
	if m == nil {
		return gerror.New("authentication provider manager is nil")
	}
	pluginID = strings.TrimSpace(pluginID)
	normalized, err := NormalizeScheme(scheme)
	if err != nil {
		return err
	}
	if pluginID == "" || factory == nil {
		return gerror.Newf("authentication provider replacement is invalid: scheme=%s plugin=%s", normalized, pluginID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	current := m.loadSnapshot()
	existing, ok := current.providers[normalized]
	if !ok {
		return gerror.Newf("authentication scheme is not registered: scheme=%s", normalized)
	}
	managed, ok := existing.(*managedProvider)
	if !ok || managed.pluginID != pluginID {
		return gerror.Newf("authentication scheme owner mismatch: scheme=%s plugin=%s", normalized, pluginID)
	}
	providers := cloneProviders(current.providers)
	providers[normalized] = &managedProvider{
		pluginID:   pluginID,
		factory:    factory,
		enablement: m.enablement,
		envFactory: m.envFactory,
	}
	m.snapshot.Store(&providerSnapshot{providers: providers})
	return nil
}

// Provider returns the current provider wrapper for one normalized scheme.
// Unknown schemes return found=false and must be rejected by the dispatcher.
func (m *Manager) Provider(scheme string) (provider Provider, found bool) {
	if m == nil {
		return nil, false
	}
	normalized, err := NormalizeScheme(scheme)
	if err != nil {
		return nil, false
	}
	provider, found = m.loadSnapshot().providers[normalized]
	return provider, found
}

// Authenticate resolves the exact scheme from the current atomic snapshot and
// delegates once. Unknown or disabled schemes fail closed without fallback.
func (m *Manager) Authenticate(
	ctx context.Context,
	scheme string,
	request authcap.AuthenticationRequest,
) (authcap.AuthenticationResult, error) {
	provider, ok := m.Provider(scheme)
	if !ok || provider == nil {
		return authcap.AuthenticationResult{}, ErrProviderUnavailable
	}
	return provider.Authenticate(ctx, request)
}

// Authenticate checks current enablement, lazily constructs the provider from
// the shared environment, delegates once, and checks enablement again before
// returning a trusted result.
func (p *managedProvider) Authenticate(
	ctx context.Context,
	request authcap.AuthenticationRequest,
) (authcap.AuthenticationResult, error) {
	if p == nil || p.enablement == nil || !p.enablement.IsProviderEnabled(ctx, p.pluginID) {
		return authcap.AuthenticationResult{}, ErrProviderUnavailable
	}
	provider, err := p.resolve(ctx)
	if err != nil {
		return authcap.AuthenticationResult{}, err
	}
	result, err := provider.Authenticate(ctx, request)
	if err != nil {
		return authcap.AuthenticationResult{}, err
	}
	if !p.enablement.IsProviderEnabled(ctx, p.pluginID) {
		return authcap.AuthenticationResult{}, ErrProviderUnavailable
	}
	return result, nil
}

// resolve lazily creates and caches the provider for this immutable managed
// wrapper. Factory failures are not cached so repaired dependencies can recover.
func (p *managedProvider) resolve(ctx context.Context) (Provider, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.provider != nil {
		return p.provider, nil
	}
	env := p.envFactory(ctx, p.pluginID)
	provider, err := p.factory(ctx, env)
	if err != nil {
		return nil, gerror.Wrapf(err, "create authentication provider failed: plugin=%s", p.pluginID)
	}
	if provider == nil {
		return nil, gerror.Newf("authentication provider factory returned nil: plugin=%s", p.pluginID)
	}
	p.provider = provider
	return provider, nil
}

// loadSnapshot returns the current immutable provider directory.
func (m *Manager) loadSnapshot() *providerSnapshot {
	if snapshot := m.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return &providerSnapshot{providers: map[string]Provider{}}
}

// cloneProviders detaches a provider lookup map before atomic publication.
func cloneProviders(source map[string]Provider) map[string]Provider {
	cloned := make(map[string]Provider, len(source)+1)
	for scheme, provider := range source {
		cloned[scheme] = provider
	}
	return cloned
}
