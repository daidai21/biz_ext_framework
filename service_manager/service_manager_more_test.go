package service_manager

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_component"
	"github.com/daidai21/biz_ext_framework/biz_ctx"
	"github.com/daidai21/biz_ext_framework/biz_process"
	"github.com/daidai21/biz_ext_framework/ext_interceptor"
	"github.com/daidai21/biz_ext_framework/ext_process"
	"github.com/daidai21/biz_ext_framework/ext_spi"
)

func TestComponentContainerWrapperMethods(t *testing.T) {
	ctxContainer := NewCtxContainer()
	if _, err := ctxContainer.Create("s1"); err != nil {
		t.Fatalf("create session failed: %v", err)
	}

	container := NewComponentContainer(ctxContainer)
	if err := container.RegisterAny("svc", biz_component.ServiceScope, func(context.Context, biz_component.Resolver) (any, error) {
		return "service-value", nil
	}); err != nil {
		t.Fatalf("register service failed: %v", err)
	}
	if err := container.RegisterAny("sess", biz_component.SessionScope, func(ctx context.Context, _ biz_component.Resolver) (any, error) {
		session, ok := ctxContainer.SessionFromContext(ctx)
		if !ok {
			t.Fatal("expected session in context")
		}
		return "session:" + session.BizSessionId(), nil
	}); err != nil {
		t.Fatalf("register session failed: %v", err)
	}

	if value, err := container.ResolveAny(context.Background(), "svc"); err != nil || value.(string) != "service-value" {
		t.Fatalf("unexpected service resolve result: %v %v", value, err)
	}
	if value, err := container.ResolveAnyInSession(context.Background(), "s1", "sess"); err != nil || value.(string) != "session:s1" {
		t.Fatalf("unexpected session resolve result: %v %v", value, err)
	}
	if _, err := container.ResolveAnyInSession(context.Background(), "missing", "sess"); !errors.Is(err, ErrContainerNotFound) {
		t.Fatalf("expected ErrContainerNotFound, got %v", err)
	}

	if value, ok := container.ServiceObject("svc"); !ok || value.(string) != "service-value" {
		t.Fatalf("unexpected service object: %v %v", value, ok)
	}
	if value, ok := container.SessionObject("s1", "sess"); !ok || value.(string) != "session:s1" {
		t.Fatalf("unexpected session object: %v %v", value, ok)
	}
	if len(container.ServiceObjects()) != 1 || len(container.SessionObjects("s1")) != 1 {
		t.Fatalf("unexpected object snapshots")
	}
	if len(container.ServiceNames()) != 1 || container.ServiceNames()[0] != "svc" {
		t.Fatalf("unexpected service names: %v", container.ServiceNames())
	}
	if len(container.SessionNames("s1")) != 1 || container.SessionNames("s1")[0] != "sess" {
		t.Fatalf("unexpected session names: %v", container.SessionNames("s1"))
	}

	container.DeleteService("svc")
	container.DeleteSessionObject("s1", "sess")
	container.ClearSession("s1")
	if _, ok := container.ServiceObject("svc"); ok {
		t.Fatal("expected deleted service object")
	}
	if _, ok := container.SessionObject("s1", "sess"); ok {
		t.Fatal("expected deleted session object")
	}
}

func TestCtxContainerManagementHelpers(t *testing.T) {
	container := NewCtxContainer()
	sessionB := biz_ctx.NewBizSession("b")
	sessionA := biz_ctx.NewBizSession("a")
	if err := container.Register(sessionB); err != nil {
		t.Fatalf("register b failed: %v", err)
	}
	if err := container.Register(sessionA); err != nil {
		t.Fatalf("register a failed: %v", err)
	}

	if ids := container.SessionIDs(); len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Fatalf("unexpected session ids: %v", ids)
	}
	if _, ok := container.Get("a"); !ok {
		t.Fatal("expected session a to exist")
	}
	container.Remove("a")
	if _, ok := container.Get("a"); ok {
		t.Fatal("expected removed session a to be absent")
	}
	if _, err := container.WithSession(context.Background(), "a"); !errors.Is(err, ErrContainerNotFound) {
		t.Fatalf("expected ErrContainerNotFound, got %v", err)
	}
}

func TestIdentityContainerHelpers(t *testing.T) {
	container, err := NewIdentityContainer("SELLER.SHOP", "SELLER.BIZ")
	if err != nil {
		t.Fatalf("new identity container failed: %v", err)
	}
	if scopes := container.Scopes(); len(scopes) != 2 || scopes[0] != "SELLER.BIZ" || scopes[1] != "SELLER.SHOP" {
		t.Fatalf("unexpected scopes: %v", scopes)
	}
	if container.IsIdentityAllowed(nil) {
		t.Fatal("expected nil identity to be rejected")
	}
}

func TestProcessContainerManagement(t *testing.T) {
	container := NewProcessContainer()
	if err := container.Register("", biz_process.Process{Layers: []biz_process.ProcessLayer{{Name: "x", Nodes: []biz_process.ProcessNode{biz_process.Task{Name: "n", Task: func(context.Context) error { return nil }}}}}}); !errors.Is(err, ErrInvalidProcessName) {
		t.Fatalf("expected ErrInvalidProcessName, got %v", err)
	}
	if err := container.Register("empty", biz_process.Process{}); !errors.Is(err, biz_process.ErrEmptyProcess) {
		t.Fatalf("expected ErrEmptyProcess, got %v", err)
	}

	p1 := biz_process.Process{Layers: []biz_process.ProcessLayer{{Name: "l1", Nodes: []biz_process.ProcessNode{biz_process.Task{Name: "n1", Task: func(context.Context) error { return nil }}}}}}
	p2 := biz_process.Process{Layers: []biz_process.ProcessLayer{{Name: "l2", Nodes: []biz_process.ProcessNode{biz_process.Task{Name: "n2", Task: func(context.Context) error { return nil }}}}}}
	if err := container.Register("b", p1); err != nil {
		t.Fatalf("register b failed: %v", err)
	}
	if err := container.Register("a", p2); err != nil {
		t.Fatalf("register a failed: %v", err)
	}
	if names := container.Names(); len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("unexpected process names: %v", names)
	}
	container.Unregister("a")
	if _, ok := container.Get("a"); ok {
		t.Fatal("expected removed process to be absent")
	}
}

func TestExtProcessContainerManagement(t *testing.T) {
	template := ext_process.NewTemplate(func(context.Context, testExtProcess, string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testExtProcess, input string) (string, bool, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewExtProcessContainer[testExtProcess, string, string](template)
	if err != nil {
		t.Fatalf("new ext process container failed: %v", err)
	}
	if err := container.Replace("b", []testExtProcess{testExtProcessImpl{value: "b"}}); err != nil {
		t.Fatalf("replace b failed: %v", err)
	}
	if err := container.Replace("a", []testExtProcess{testExtProcessImpl{value: "a"}}); err != nil {
		t.Fatalf("replace a failed: %v", err)
	}
	if defs := container.Definitions(); len(defs) != 2 || defs[0] != "a" || defs[1] != "b" {
		t.Fatalf("unexpected ext process definitions: %v", defs)
	}
	container.Remove("a")
	if len(container.Implementations("a")) != 0 {
		t.Fatal("expected removed ext process implementations to be empty")
	}
	if _, err := container.Execute(context.Background(), "", "input", ext_process.Serial); !errors.Is(err, ErrInvalidExtProcessDefinition) {
		t.Fatalf("expected ErrInvalidExtProcessDefinition, got %v", err)
	}

	nilTemplateContainer := &ExtProcessContainer[testExtProcess, string, string]{}
	if _, err := nilTemplateContainer.Execute(context.Background(), "audit", "input", ext_process.Serial); !errors.Is(err, ErrNilExtProcessTemplate) {
		t.Fatalf("expected ErrNilExtProcessTemplate, got %v", err)
	}
}

func TestSPIContainerManagementHelpers(t *testing.T) {
	template := ext_spi.NewTemplate(func(context.Context, testSPI, string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testSPI, input string) (string, error) {
		return impl.Handle(ctx, input)
	})
	container, err := NewSPIContainer[testSPI, string, string](template)
	if err != nil {
		t.Fatalf("new spi container failed: %v", err)
	}
	if err := container.Replace("b", []testSPI{testSPIImpl{value: "b"}}); err != nil {
		t.Fatalf("replace b failed: %v", err)
	}
	if err := container.Replace("a", []testSPI{testSPIImpl{value: "a"}}); err != nil {
		t.Fatalf("replace a failed: %v", err)
	}
	if defs := container.Definitions(); len(defs) != 2 || defs[0] != "a" || defs[1] != "b" {
		t.Fatalf("unexpected spi definitions: %v", defs)
	}
	container.Remove("a")
	if len(container.Implementations("a")) != 0 {
		t.Fatal("expected removed spi implementations to be empty")
	}
	if _, err := container.Execute(context.Background(), "", "input", ext_spi.All); !errors.Is(err, ErrInvalidSPIDefinition) {
		t.Fatalf("expected ErrInvalidSPIDefinition, got %v", err)
	}

	nilTemplateContainer := &SPIContainer[testSPI, string, string]{}
	if _, err := nilTemplateContainer.Execute(context.Background(), "audit", "input", ext_spi.All); !errors.Is(err, ErrNilSPITemplate) {
		t.Fatalf("expected ErrNilSPITemplate, got %v", err)
	}
}

func TestInterceptorContainerManagementHelpers(t *testing.T) {
	template := ext_interceptor.NewTemplate(func(context.Context, testInterceptorSPI, string) (bool, error) {
		return true, nil
	}, func(ctx context.Context, impl testInterceptorSPI, input string, next ext_interceptor.Handler[string, string]) (string, error) {
		return impl.Handle(ctx, input, next)
	})
	container, err := NewInterceptorContainer[testInterceptorSPI, string, string](template)
	if err != nil {
		t.Fatalf("new interceptor container failed: %v", err)
	}
	if err := container.Replace("b", []testInterceptorSPI{testInterceptorImpl{name: "b"}}); err != nil {
		t.Fatalf("replace b failed: %v", err)
	}
	if err := container.Replace("a", []testInterceptorSPI{testInterceptorImpl{name: "a"}}); err != nil {
		t.Fatalf("replace a failed: %v", err)
	}
	if defs := container.Definitions(); len(defs) != 2 || defs[0] != "a" || defs[1] != "b" {
		t.Fatalf("unexpected interceptor definitions: %v", defs)
	}
	container.Remove("a")
	if len(container.Interceptors("a")) != 0 {
		t.Fatal("expected removed interceptors to be empty")
	}
	if _, err := container.Execute(context.Background(), "", "input", func(context.Context, string) (string, error) { return "", nil }); !errors.Is(err, ErrInvalidInterceptorDefinition) {
		t.Fatalf("expected ErrInvalidInterceptorDefinition, got %v", err)
	}

	nilTemplateContainer := &InterceptorContainer[testInterceptorSPI, string, string]{}
	if _, err := nilTemplateContainer.Execute(context.Background(), "rpc", "input", func(context.Context, string) (string, error) { return "", nil }); !errors.Is(err, ErrNilInterceptorTemplate) {
		t.Fatalf("expected ErrNilInterceptorTemplate, got %v", err)
	}
}

func TestServiceManagerBuilderValidationPaths(t *testing.T) {
	if _, err := NewServiceManagerBuilder("svc").WithIdentityContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for identity container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithProcessContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for process container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithModelContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for model container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithCtxContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for ctx container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithComponentContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for component container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithObservationContainer(nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for observation container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithContainer("", struct{}{}).Build(); !errors.Is(err, ErrInvalidContainerName) {
		t.Fatalf("expected ErrInvalidContainerName, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithContainer("x", nil).Build(); !errors.Is(err, ErrNilContainer) {
		t.Fatalf("expected ErrNilContainer for custom container, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithStartupCheck(nil).Build(); !errors.Is(err, ErrNilStartupCheck) {
		t.Fatalf("expected ErrNilStartupCheck, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithLifecycle("", &fakeLifecycle{}).Build(); !errors.Is(err, ErrInvalidLifecycleName) {
		t.Fatalf("expected ErrInvalidLifecycleName, got %v", err)
	}
	if _, err := NewServiceManagerBuilder("svc").WithLifecycle("x", nil).Build(); !errors.Is(err, ErrNilLifecycle) {
		t.Fatalf("expected ErrNilLifecycle, got %v", err)
	}

	if _, err := NewServiceManagerBuilder("svc").WithIdentityScopes("seller.shop").Build(); err == nil {
		t.Fatal("expected invalid identity scope error")
	}
	if _, err := NewServiceManagerBuilder("svc").WithProcess("bad", biz_process.Process{}).Build(); err == nil {
		t.Fatal("expected invalid process registration error")
	}
	if _, err := NewServiceManagerBuilder("svc").WithModelWhitelist("invalid", "a").Build(); err == nil {
		t.Fatal("expected invalid model whitelist error")
	}
	if _, err := NewServiceManagerBuilder("svc").WithContainer(IdentityContainerName, struct{}{}).Build(); err == nil {
		t.Fatal("expected duplicate container error")
	}
}

func TestServiceManagerMustContainerAndCheckFailures(t *testing.T) {
	manager, err := NewServiceManagerBuilder("svc").
		WithStartupCheck(func(context.Context, *ServiceManager) error {
			return errors.New("check failed")
		}).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if _, err := manager.MustContainer("missing"); !errors.Is(err, ErrContainerNotFound) {
		t.Fatalf("expected ErrContainerNotFound, got %v", err)
	}
	if err := manager.Check(context.Background()); err == nil || !strings.Contains(err.Error(), "startup check[0] failed") {
		t.Fatalf("expected wrapped startup check error, got %v", err)
	}
	if err := manager.Start(context.Background()); err == nil || !strings.Contains(err.Error(), "startup check[0] failed") {
		t.Fatalf("expected start check failure, got %v", err)
	}
	if manager.State() != ServiceManagerStateReady {
		t.Fatalf("expected READY after failed start, got %s", manager.State())
	}
}

func TestServiceManagerStopErrorAndBusyStop(t *testing.T) {
	stopStarted := make(chan struct{})
	stopRelease := make(chan struct{})

	manager, err := NewServiceManagerBuilder("svc").
		WithLifecycle("hold", &fakeLifecycle{
			stopFn: func() error {
				close(stopStarted)
				<-stopRelease
				return errors.New("stop boom")
			},
		}).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("start failed: %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- manager.Stop(context.Background())
	}()

	<-stopStarted
	if err := manager.Stop(context.Background()); !errors.Is(err, ErrServiceManagerBusy) {
		t.Fatalf("expected ErrServiceManagerBusy, got %v", err)
	}
	close(stopRelease)

	if err := <-errCh; err == nil || !strings.Contains(err.Error(), "stop boom") {
		t.Fatalf("expected stop error, got %v", err)
	}
	if manager.State() != ServiceManagerStateStopped {
		t.Fatalf("expected STOPPED after stop error, got %s", manager.State())
	}
}
