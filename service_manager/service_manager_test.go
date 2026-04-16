package service_manager

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/daidai21/biz_ext_framework/biz_process"
)

type fakeLifecycle struct {
	startFn func() error
	stopFn  func() error
}

func (l *fakeLifecycle) Start(ctx context.Context) error {
	if l.startFn != nil {
		return l.startFn()
	}
	return nil
}

func (l *fakeLifecycle) Stop(ctx context.Context) error {
	if l.stopFn != nil {
		return l.stopFn()
	}
	return nil
}

func TestServiceManagerBuilderBuild(t *testing.T) {
	var startupChecked bool
	customSPI := NewSPIContainer[string]()

	manager, err := NewServiceManagerBuilder("order-service").
		WithIdentityScopes("SELLER.SHOP").
		WithProcess("order_flow", biz_process.Process{
			Layers: []biz_process.ProcessLayer{
				{
					Name: "prepare",
					Nodes: []biz_process.ProcessNode{
						biz_process.TaskProcessNode{Name: "prepare", Task: func(ctx context.Context) error { return nil }},
					},
				},
			},
		}).
		WithModelWhitelist("psm.order#CreateOrder", "user").
		WithContainer("spi_container", customSPI).
		WithStartupCheck(func(ctx context.Context, manager *ServiceManager) error {
			startupChecked = true
			if !manager.IdentityContainer().IsAllowed("SELLER.SHOP.OPERATOR") {
				return errors.New("identity scope missing")
			}
			if _, ok := manager.ProcessContainer().Get("order_flow"); !ok {
				return errors.New("process missing")
			}
			return nil
		}).
		Build()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if manager.Name() != "order-service" {
		t.Fatalf("unexpected manager name: %s", manager.Name())
	}
	if manager.State() != ServiceManagerStateReady {
		t.Fatalf("expected READY, got %s", manager.State())
	}
	if !manager.IdentityContainer().IsAllowed("SELLER.SHOP.OPERATOR") {
		t.Fatalf("expected identity scope initialized")
	}
	if _, ok := manager.ProcessContainer().Get("order_flow"); !ok {
		t.Fatalf("expected process initialized")
	}
	if container, ok := manager.Container("spi_container"); !ok || container != customSPI {
		t.Fatalf("expected custom spi container registered")
	}
	if whitelist := manager.ModelContainer().Whitelist("psm.order#CreateOrder"); len(whitelist) != 1 || whitelist[0] != "user" {
		t.Fatalf("unexpected whitelist: %v", whitelist)
	}
	if err := manager.Check(context.Background()); err != nil {
		t.Fatalf("expected check success, got %v", err)
	}
	if !startupChecked {
		t.Fatalf("expected startup check executed")
	}
}

func TestServiceManagerLifecycle(t *testing.T) {
	var order []string
	manager, err := NewServiceManagerBuilder("order-service").
		WithLifecycle("first", &fakeLifecycle{
			startFn: func() error {
				order = append(order, "start:first")
				return nil
			},
			stopFn: func() error {
				order = append(order, "stop:first")
				return nil
			},
		}).
		WithLifecycle("second", &fakeLifecycle{
			startFn: func() error {
				order = append(order, "start:second")
				return nil
			},
			stopFn: func() error {
				order = append(order, "stop:second")
				return nil
			},
		}).
		Build()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("expected start success, got %v", err)
	}
	if manager.State() != ServiceManagerStateStarted {
		t.Fatalf("expected STARTED, got %s", manager.State())
	}
	if err := manager.Stop(context.Background()); err != nil {
		t.Fatalf("expected stop success, got %v", err)
	}
	if manager.State() != ServiceManagerStateStopped {
		t.Fatalf("expected STOPPED, got %s", manager.State())
	}

	got := strings.Join(order, ",")
	if got != "start:first,start:second,stop:second,stop:first" {
		t.Fatalf("unexpected lifecycle order: %s", got)
	}
}

func TestServiceManagerStartRollbackOnFailure(t *testing.T) {
	var order []string
	manager, err := NewServiceManagerBuilder("order-service").
		WithLifecycle("first", &fakeLifecycle{
			startFn: func() error {
				order = append(order, "start:first")
				return nil
			},
			stopFn: func() error {
				order = append(order, "stop:first")
				return nil
			},
		}).
		WithLifecycle("bad", &fakeLifecycle{
			startFn: func() error {
				order = append(order, "start:bad")
				return errors.New("boom")
			},
		}).
		Build()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	err = manager.Start(context.Background())
	if err == nil {
		t.Fatalf("expected start fail")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected error contains boom, got %v", err)
	}
	if manager.State() != ServiceManagerStateReady {
		t.Fatalf("expected READY after failed start, got %s", manager.State())
	}

	got := strings.Join(order, ",")
	if got != "start:first,start:bad,stop:first" {
		t.Fatalf("unexpected rollback order: %s", got)
	}
}

func TestServiceManagerStateErrors(t *testing.T) {
	manager, err := NewServiceManagerBuilder("order-service").Build()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if err := manager.Stop(context.Background()); !errors.Is(err, ErrServiceManagerNotStarted) {
		t.Fatalf("expected ErrServiceManagerNotStarted, got %v", err)
	}
	if err := manager.Start(context.Background()); err != nil {
		t.Fatalf("expected start success, got %v", err)
	}
	if err := manager.Start(context.Background()); !errors.Is(err, ErrServiceManagerStarted) {
		t.Fatalf("expected ErrServiceManagerStarted, got %v", err)
	}
	if err := manager.Stop(context.Background()); err != nil {
		t.Fatalf("expected stop success, got %v", err)
	}
	if err := manager.Stop(context.Background()); !errors.Is(err, ErrServiceManagerStopped) {
		t.Fatalf("expected ErrServiceManagerStopped, got %v", err)
	}
}

func TestServiceManagerConcurrentStartBusy(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})

	manager, err := NewServiceManagerBuilder("order-service").
		WithLifecycle("first", &fakeLifecycle{
			startFn: func() error {
				close(started)
				<-release
				return nil
			},
		}).
		Build()
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- manager.Start(context.Background())
	}()

	<-started
	if err := manager.Start(context.Background()); !errors.Is(err, ErrServiceManagerBusy) {
		t.Fatalf("expected ErrServiceManagerBusy, got %v", err)
	}

	close(release)
	if err := <-errCh; err != nil {
		t.Fatalf("expected first start success, got %v", err)
	}
}

func TestServiceManagerBuilderInvalidName(t *testing.T) {
	_, err := NewServiceManagerBuilder("").Build()
	if !errors.Is(err, ErrInvalidServiceManagerName) {
		t.Fatalf("expected ErrInvalidServiceManagerName, got %v", err)
	}
}
