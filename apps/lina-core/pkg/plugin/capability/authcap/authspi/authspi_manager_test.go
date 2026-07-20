// This file verifies authentication provider registration, lifecycle gating,
// dependency failures, and atomic upgrade switching.

package authspi

import (
	"context"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/pkg/plugin/capability/authcap"
)

// testEnablement provides mutable provider enablement for one self-contained test.
type testEnablement struct {
	enabled map[string]bool
}

// IsProviderEnabled returns the current test state for pluginID.
func (e *testEnablement) IsProviderEnabled(_ context.Context, pluginID string) bool {
	return e != nil && e.enabled[pluginID]
}

// testProvider returns one stable machine subject for manager assertions.
type testProvider struct {
	subject string
}

// Authenticate returns a machine actor carrying the configured subject.
func (p *testProvider) Authenticate(
	context.Context,
	authcap.AuthenticationRequest,
) (authcap.AuthenticationResult, error) {
	return authcap.AuthenticationResult{Actor: authcap.Actor{
		Kind:         authcap.ActorKindMachine,
		SubjectID:    p.subject,
		CredentialID: "credential-1",
	}}, nil
}

// TestManagerRejectsMissingDependenciesAndDuplicateSchemes verifies startup
// dependency requirements and global case-insensitive scheme uniqueness.
func TestManagerRejectsMissingDependenciesAndDuplicateSchemes(t *testing.T) {
	t.Parallel()
	enablement := &testEnablement{enabled: map[string]bool{"plugin-a": true, "plugin-b": true}}
	envFactory := func(_ context.Context, pluginID string) ProviderEnv { return ProviderEnv{PluginID: pluginID} }
	if _, err := NewManager(nil, envFactory); err == nil {
		t.Fatal("expected missing enablement reader to fail")
	}
	if _, err := NewManager(enablement, nil); err == nil {
		t.Fatal("expected missing environment factory to fail")
	}
	manager, err := NewManager(enablement, envFactory)
	if err != nil {
		t.Fatalf("expected manager construction to succeed: %v", err)
	}
	factory := func(context.Context, ProviderEnv) (Provider, error) { return &testProvider{subject: "a"}, nil }
	if err = manager.RegisterFactory("plugin-a", "lina-hmac-sha256", factory); err != nil {
		t.Fatalf("expected first scheme registration to succeed: %v", err)
	}
	if err = manager.RegisterFactory("plugin-b", "LINA-HMAC-SHA256", factory); err == nil {
		t.Fatal("expected duplicate normalized scheme registration to fail")
	}
}

// TestManagerFollowsEnablementAndRecoversFactoryFailures verifies disabled
// providers and missing runtime dependencies fail closed without poisoning
// later recovery.
func TestManagerFollowsEnablementAndRecoversFactoryFailures(t *testing.T) {
	t.Parallel()
	enablement := &testEnablement{enabled: map[string]bool{"plugin-a": false}}
	ready := false
	manager, err := NewManager(enablement, func(_ context.Context, pluginID string) ProviderEnv {
		return ProviderEnv{PluginID: pluginID}
	})
	if err != nil {
		t.Fatalf("expected manager construction to succeed: %v", err)
	}
	err = manager.RegisterFactory("plugin-a", "LINA-HMAC-SHA256", func(context.Context, ProviderEnv) (Provider, error) {
		if !ready {
			return nil, gerror.New("required dependency unavailable")
		}
		return &testProvider{subject: "recovered"}, nil
	})
	if err != nil {
		t.Fatalf("expected scheme registration to succeed: %v", err)
	}
	request := newTestAuthenticationRequest()
	if _, err = manager.Authenticate(context.Background(), request.Scheme(), request); err == nil {
		t.Fatal("expected disabled provider to fail closed")
	}
	enablement.enabled["plugin-a"] = true
	if _, err = manager.Authenticate(context.Background(), request.Scheme(), request); err == nil {
		t.Fatal("expected missing factory dependency to fail closed")
	}
	ready = true
	result, err := manager.Authenticate(context.Background(), request.Scheme(), request)
	if err != nil {
		t.Fatalf("expected provider to recover after dependency repair: %v", err)
	}
	if result.Actor.SubjectID != "recovered" {
		t.Fatalf("unexpected recovered subject: %s", result.Actor.SubjectID)
	}
	enablement.enabled["plugin-a"] = false
	if _, err = manager.Authenticate(context.Background(), request.Scheme(), request); err == nil {
		t.Fatal("expected provider disable to take effect after prior success")
	}
}

// TestManagerAtomicallySwitchesSameOwnerFactory verifies future lookups use a
// replacement wrapper while an existing provider reference remains isolated.
func TestManagerAtomicallySwitchesSameOwnerFactory(t *testing.T) {
	t.Parallel()
	enablement := &testEnablement{enabled: map[string]bool{"plugin-a": true}}
	manager, err := NewManager(enablement, func(_ context.Context, pluginID string) ProviderEnv {
		return ProviderEnv{PluginID: pluginID}
	})
	if err != nil {
		t.Fatalf("expected manager construction to succeed: %v", err)
	}
	oldFactory := func(context.Context, ProviderEnv) (Provider, error) { return &testProvider{subject: "old"}, nil }
	newFactory := func(context.Context, ProviderEnv) (Provider, error) { return &testProvider{subject: "new"}, nil }
	if err = manager.RegisterFactory("plugin-a", "LINA-HMAC-SHA256", oldFactory); err != nil {
		t.Fatalf("expected scheme registration to succeed: %v", err)
	}
	request := newTestAuthenticationRequest()
	result, err := manager.Authenticate(context.Background(), request.Scheme(), request)
	if err != nil || result.Actor.SubjectID != "old" {
		t.Fatalf("expected old provider before replacement, result=%#v err=%v", result, err)
	}
	if err = manager.ReplaceFactory("plugin-a", request.Scheme(), newFactory); err != nil {
		t.Fatalf("expected same-owner factory replacement to succeed: %v", err)
	}
	result, err = manager.Authenticate(context.Background(), request.Scheme(), request)
	if err != nil || result.Actor.SubjectID != "new" {
		t.Fatalf("expected new provider after replacement, result=%#v err=%v", result, err)
	}
	if err = manager.ReplaceFactory("plugin-b", request.Scheme(), oldFactory); err == nil {
		t.Fatal("expected cross-owner factory replacement to fail")
	}
}

// newTestAuthenticationRequest creates one request projection for manager tests.
func newTestAuthenticationRequest() authcap.AuthenticationRequest {
	return authcap.NewAuthenticationRequest(
		"LINA-HMAC-SHA256",
		"Credential=test,Signature=test",
		"GET",
		"/machine/test",
		nil,
		nil,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		"127.0.0.1",
	)
}
