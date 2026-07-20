// This file verifies the real GoFrame middleware chain for precise user and
// machine authentication dispatch, authorization, early exits, and tenant trust.

package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"

	"lina-core/internal/model"
	"lina-core/internal/service/auth"
	"lina-core/internal/service/bizctx"
	"lina-core/internal/service/role"
	"lina-core/internal/service/session"
	"lina-core/pkg/plugin/capability/authcap"
)

type machineAuthAllowedReq struct {
	g.Meta `path:"/machine" method:"get" operation:"records.list" resource:"records" action:"read" actors:"user,machine" permission:"records:view"`
}

type machineAuthUserOnlyReq struct {
	g.Meta `path:"/user-only" method:"get" permission:"records:view"`
}

type machineAuthWriteReq struct {
	g.Meta `path:"/machine-write" method:"post" operation:"records.create" resource:"records" action:"write" actors:"machine"`
}

type machineAuthTestRes struct {
	TenantID int `json:"tenantId"`
}

type machineAuthTestController struct {
	bizCtx bizctx.Service
	called *atomic.Int32
	tenant *atomic.Int32
}

func (c *machineAuthTestController) Allowed(ctx context.Context, _ *machineAuthAllowedReq) (*machineAuthTestRes, error) {
	c.called.Add(1)
	tenantID := c.bizCtx.Get(ctx).TenantId
	c.tenant.Store(int32(tenantID))
	return &machineAuthTestRes{TenantID: tenantID}, nil
}

func (c *machineAuthTestController) UserOnly(ctx context.Context, _ *machineAuthUserOnlyReq) (*machineAuthTestRes, error) {
	c.called.Add(1)
	tenantID := c.bizCtx.Get(ctx).TenantId
	c.tenant.Store(int32(tenantID))
	return &machineAuthTestRes{TenantID: tenantID}, nil
}

func (c *machineAuthTestController) Write(ctx context.Context, _ *machineAuthWriteReq) (*machineAuthTestRes, error) {
	c.called.Add(1)
	tenantID := c.bizCtx.Get(ctx).TenantId
	c.tenant.Store(int32(tenantID))
	return &machineAuthTestRes{TenantID: tenantID}, nil
}

type machineAuthDispatcherStub struct {
	result authcap.AuthenticationResult
	err    error
	check  func(authcap.AuthenticationRequest) error
	calls  atomic.Int32
}

func (s *machineAuthDispatcherStub) Authenticate(
	_ context.Context,
	_ string,
	request authcap.AuthenticationRequest,
) (authcap.AuthenticationResult, error) {
	s.calls.Add(1)
	if s.check != nil {
		if err := s.check(request); err != nil {
			return authcap.AuthenticationResult{}, err
		}
	}
	return s.result, s.err
}

type middlewareAuthServiceStub struct {
	auth.Service
	claims *auth.Claims
	err    error
}

func (s middlewareAuthServiceStub) AuthenticateAccessToken(context.Context, string) (*auth.Claims, error) {
	return s.claims, s.err
}

func (middlewareAuthServiceStub) SessionStore() session.Store { return nil }

type middlewareRoleServiceStub struct {
	role.Service
	access *role.UserAccessContext
}

func (s middlewareRoleServiceStub) GetUserAccessContext(context.Context, int) (*role.UserAccessContext, error) {
	return s.access, nil
}

func TestMachineAuthenticationMiddlewareChain(t *testing.T) {
	testCases := []struct {
		name             string
		method           string
		path             string
		result           authcap.AuthenticationResult
		providerErr      error
		expectedStatus   int
		expectedCalls    int32
		expectedTenantID int
	}{
		{
			name: "allowed operation and resource",
			path: "/machine",
			result: machineAuthenticationResult(
				[]authcap.OperationCode{"records.list"},
				[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeRead}},
			),
			expectedStatus: http.StatusOK, expectedCalls: 1, expectedTenantID: 77,
		},
		{
			name: "user-only route defaults machine denied",
			path: "/user-only",
			result: machineAuthenticationResult(
				[]authcap.OperationCode{"records.list"},
				[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeRead}},
			),
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "operation denied",
			path: "/machine",
			result: machineAuthenticationResult(
				nil,
				[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeRead}},
			),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "resource denied",
			path:           "/machine",
			result:         machineAuthenticationResult([]authcap.OperationCode{"records.list"}, nil),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:   "write operation and resource allowed",
			method: http.MethodPost,
			path:   "/machine-write",
			result: machineAuthenticationResult(
				[]authcap.OperationCode{"records.create"},
				[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeWrite}},
			),
			expectedStatus: http.StatusOK, expectedCalls: 1, expectedTenantID: 77,
		},
		{
			name:   "write resource denied by read-only policy",
			method: http.MethodPost,
			path:   "/machine-write",
			result: machineAuthenticationResult(
				[]authcap.OperationCode{"records.create"},
				[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeRead}},
			),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "provider authentication failure",
			path:           "/machine",
			providerErr:    gerror.New("invalid credential"),
			expectedStatus: http.StatusUnauthorized,
		},
	}
	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			dispatcher := &machineAuthDispatcherStub{result: testCase.result, err: testCase.providerErr}
			server, called, observedTenant := startMachineAuthMiddlewareServer(t, nil, nil, dispatcher)
			method := testCase.method
			if method == "" {
				method = http.MethodGet
			}
			request, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:%d%s", server.GetListenedPort(), testCase.path), nil)
			if err != nil {
				t.Fatalf("create request: %v", err)
			}
			request.Header.Set("Authorization", "TEST-MACHINE credential")
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("execute request: %v", err)
			}
			defer response.Body.Close()
			if response.StatusCode != testCase.expectedStatus {
				t.Fatalf("expected status %d, got %d", testCase.expectedStatus, response.StatusCode)
			}
			if called.Load() != testCase.expectedCalls {
				t.Fatalf("expected controller calls %d, got %d", testCase.expectedCalls, called.Load())
			}
			if dispatcher.calls.Load() != 1 {
				t.Fatalf("expected exact provider dispatch once, got %d", dispatcher.calls.Load())
			}
			if testCase.expectedTenantID > 0 && int(observedTenant.Load()) != testCase.expectedTenantID {
				t.Fatalf("expected trusted tenant %d, got %d", testCase.expectedTenantID, observedTenant.Load())
			}
		})
	}
}

func TestSignedMachineRequestProjectionAndReplayEarlyExit(t *testing.T) {
	const (
		body      = `{"name":"signed"}`
		nonce     = "AAECAwQFBgcICQoLDA0ODw"
		timestamp = "1776756000"
	)
	digest := sha256.Sum256([]byte(body))
	bodyHash := hex.EncodeToString(digest[:])
	var (
		mu       sync.Mutex
		consumed = make(map[string]struct{})
	)
	dispatcher := &machineAuthDispatcherStub{
		result: machineAuthenticationResult(
			[]authcap.OperationCode{"records.create"},
			[]authcap.ResourcePermission{{Resource: "records", Access: authcap.AccessModeWrite}},
		),
		check: func(request authcap.AuthenticationRequest) error {
			if request.Scheme() != "LINA-HMAC-SHA256" || request.Method() != http.MethodPost {
				return gerror.New("signed request scheme or method projection mismatch")
			}
			if request.Header("X-Lina-Date") != timestamp || request.Header("X-Lina-Nonce") != nonce {
				return gerror.New("signed request header projection mismatch")
			}
			if request.BodySHA256() != bodyHash || request.Header("X-Lina-Content-SHA256") != bodyHash {
				return gerror.New("signed request body digest projection mismatch")
			}
			mu.Lock()
			defer mu.Unlock()
			if _, exists := consumed[nonce]; exists {
				return gerror.New("signed request replay rejected")
			}
			consumed[nonce] = struct{}{}
			return nil
		},
	}
	server, called, _ := startMachineAuthMiddlewareServer(t, nil, nil, dispatcher)

	for attempt, expectedStatus := range []int{http.StatusOK, http.StatusUnauthorized} {
		request, err := http.NewRequest(
			http.MethodPost,
			fmt.Sprintf("http://127.0.0.1:%d/machine-write", server.GetListenedPort()),
			strings.NewReader(body),
		)
		if err != nil {
			t.Fatalf("create signed request attempt=%d: %v", attempt+1, err)
		}
		request.Header.Set("Authorization", "LINA-HMAC-SHA256 Credential=lak_test,Signature="+strings.Repeat("ab", sha256.Size))
		request.Header.Set("X-Lina-Date", timestamp)
		request.Header.Set("X-Lina-Nonce", nonce)
		request.Header.Set("X-Lina-Content-SHA256", bodyHash)
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatalf("execute signed request attempt=%d: %v", attempt+1, err)
		}
		response.Body.Close()
		if response.StatusCode != expectedStatus {
			t.Fatalf("attempt=%d expected status %d, got %d", attempt+1, expectedStatus, response.StatusCode)
		}
	}
	if called.Load() != 1 {
		t.Fatalf("replayed request must not execute controller, calls=%d", called.Load())
	}
}

func TestBearerAuthenticationMiddlewareRegression(t *testing.T) {
	clientType, err := auth.ParseClientType("web")
	if err != nil {
		t.Fatalf("parse client type: %v", err)
	}
	dispatcher := &machineAuthDispatcherStub{}
	authSvc := middlewareAuthServiceStub{claims: &auth.Claims{
		TokenId: "token-1", UserId: 9, Username: "user", Status: 1, TenantId: 7, ClientType: clientType,
	}}
	roleSvc := middlewareRoleServiceStub{access: &role.UserAccessContext{Permissions: []string{"records:view"}}}
	server, called, _ := startMachineAuthMiddlewareServer(t, authSvc, roleSvc, dispatcher)
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://127.0.0.1:%d/user-only", server.GetListenedPort()), nil)
	if err != nil {
		t.Fatalf("create bearer request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer token")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("execute bearer request: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || called.Load() != 1 {
		t.Fatalf("expected bearer user path success, status=%d calls=%d", response.StatusCode, called.Load())
	}
	if dispatcher.calls.Load() != 0 {
		t.Fatalf("expected Bearer to bypass machine dispatcher, got %d calls", dispatcher.calls.Load())
	}
}

func startMachineAuthMiddlewareServer(
	t *testing.T,
	authSvc auth.Service,
	roleSvc role.Service,
	dispatcher *machineAuthDispatcherStub,
) (*ghttp.Server, *atomic.Int32, *atomic.Int32) {
	t.Helper()
	server := ghttp.GetServer("middleware-machine-auth-" + guid.S())
	server.SetPort(0)
	server.SetDumpRouterMap(false)
	bizCtxSvc := bizctx.New()
	called := &atomic.Int32{}
	observedTenant := &atomic.Int32{}
	service := New(authSvc, bizCtxSvc, nil, nil, roleSvc, nil, dispatcher)
	server.Group("/", func(group *ghttp.RouterGroup) {
		group.Middleware(
			ghttp.MiddlewareNeverDoneCtx,
			func(r *ghttp.Request) {
				bizCtxSvc.Init(r, &model.Context{})
				r.Middleware.Next()
			},
			service.Auth,
			service.Tenancy,
			service.Permission,
		)
		group.Bind(&machineAuthTestController{bizCtx: bizCtxSvc, called: called, tenant: observedTenant})
	})
	if err := server.Start(); err != nil {
		t.Fatalf("start middleware server: %v", err)
	}
	t.Cleanup(func() {
		if err := server.Shutdown(); err != nil {
			t.Fatalf("shutdown middleware server: %v", err)
		}
	})
	return server, called, observedTenant
}

func machineAuthenticationResult(
	operations []authcap.OperationCode,
	resources []authcap.ResourcePermission,
) authcap.AuthenticationResult {
	return authcap.AuthenticationResult{
		Actor:         authcap.Actor{Kind: authcap.ActorKindMachine, SubjectID: "client-1", CredentialID: "key-1", TenantID: 77},
		Authorization: authcap.NewAuthorizationSnapshot(operations, resources),
	}
}
