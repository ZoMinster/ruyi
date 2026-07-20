// This file synchronizes installed dynamic-plugin route authorization metadata
// into the startup-owned global catalog after lifecycle convergence.

package runtime

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"lina-core/internal/service/plugin/internal/catalog"
	"lina-core/pkg/plugin/capability/authcap"
	"lina-core/pkg/statusflag"
)

// SyncDynamicRouteAuthorizations rebuilds all installed dynamic-plugin owner
// snapshots from their active releases.
func (s *serviceImpl) SyncDynamicRouteAuthorizations(ctx context.Context) error {
	if s == nil || s.routeAuthorizations == nil {
		return gerror.New("dynamic route authorization catalog is unavailable")
	}
	registries, err := s.listRuntimeRegistries(ctx)
	if err != nil {
		return err
	}
	for _, registry := range registries {
		if registry == nil {
			continue
		}
		if err = s.syncDynamicRouteAuthorizationForPlugin(ctx, registry.PluginId); err != nil {
			return err
		}
	}
	return nil
}

// validateDynamicRouteAuthorizationCandidate checks one candidate release as
// an owner replacement without publishing it before lifecycle side effects.
func (s *serviceImpl) validateDynamicRouteAuthorizationCandidate(manifest *catalog.Manifest) error {
	if s == nil || s.routeAuthorizations == nil {
		return gerror.New("dynamic route authorization catalog is unavailable")
	}
	routes, err := dynamicRouteAuthorizations(manifest)
	if err != nil {
		return err
	}
	return s.routeAuthorizations.ValidateOwner(dynamicRouteAuthorizationOwner(manifest.ID), routes)
}

// syncDynamicRouteAuthorizationForPlugin publishes the current active route
// snapshot or removes it when the plugin is disabled, uninstalled, or absent.
func (s *serviceImpl) syncDynamicRouteAuthorizationForPlugin(ctx context.Context, pluginID string) error {
	if s == nil || s.routeAuthorizations == nil {
		return gerror.New("dynamic route authorization catalog is unavailable")
	}
	ownerKey := dynamicRouteAuthorizationOwner(pluginID)
	registry, err := s.storeSvc.GetRegistry(ctx, strings.TrimSpace(pluginID))
	if err != nil {
		return err
	}
	if registry == nil || registry.Installed != statusflag.Installed.Int() {
		s.routeAuthorizations.RemoveOwner(ownerKey)
		return nil
	}
	manifest, err := s.loadActiveManifest(ctx, registry)
	if err != nil {
		return err
	}
	routes, err := dynamicRouteAuthorizations(manifest)
	if err != nil {
		return err
	}
	return s.routeAuthorizations.ReplaceOwner(ownerKey, routes)
}

// dynamicRouteAuthorizations converts bridge contracts into the common route
// authorization projection used by host middleware.
func dynamicRouteAuthorizations(manifest *catalog.Manifest) ([]authcap.RouteAuthorization, error) {
	if manifest == nil {
		return nil, gerror.New("dynamic route authorization manifest cannot be nil")
	}
	routes := make([]authcap.RouteAuthorization, 0, len(manifest.Routes))
	for _, route := range manifest.Routes {
		if route == nil {
			return nil, gerror.Newf("dynamic route authorization contract cannot be nil: %s", manifest.ID)
		}
		item, err := authcap.ParseRouteAuthorization(
			authcap.RouteOwnerKindDynamicPlugin,
			manifest.ID,
			route.Method,
			route.Path,
			route.Operation,
			route.Resource,
			route.Action,
			route.Actors,
		)
		if err != nil {
			return nil, gerror.Wrapf(err, "dynamic route authorization metadata is invalid: %s %s", route.Method, route.Path)
		}
		routes = append(routes, item)
	}
	return routes, nil
}

func dynamicRouteAuthorizationOwner(pluginID string) string {
	return authcap.RouteAuthorizationOwnerKey(authcap.RouteOwnerKindDynamicPlugin, pluginID)
}
