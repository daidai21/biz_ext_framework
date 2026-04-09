package ext_spi

import (
	"context"
	"fmt"
	"testing"
)

type testInput struct {
	scene string
}

type testSPI interface {
	Match(ctx context.Context, input testInput) (bool, error)
	Handle(ctx context.Context, input testInput) (string, error)
}

type primarySPI struct{}

func (primarySPI) Match(_ context.Context, input testInput) (bool, error) {
	if input.scene != "ORDER" {
		return false, nil
	}
	return true, nil
}

func (primarySPI) Handle(_ context.Context, _ testInput) (string, error) {
	return "primary", nil
}

type secondarySPI struct{}

func (secondarySPI) Match(_ context.Context, input testInput) (bool, error) {
	if input.scene != "ORDER" {
		return false, nil
	}
	return true, nil
}

func (secondarySPI) Handle(_ context.Context, _ testInput) (string, error) {
	return "secondary", nil
}

type refundSPI struct{}

func (refundSPI) Match(_ context.Context, input testInput) (bool, error) {
	if input.scene != "REFUND" {
		return false, nil
	}
	return true, nil
}

func (refundSPI) Handle(_ context.Context, _ testInput) (string, error) {
	return "refund", nil
}

type errorMatchSPI struct{}

func (errorMatchSPI) Match(_ context.Context, _ testInput) (bool, error) {
	return false, fmt.Errorf("spi match failed")
}

func (errorMatchSPI) Handle(_ context.Context, _ testInput) (string, error) {
	return "", nil
}

type errorHandleSPI struct{}

func (errorHandleSPI) Match(_ context.Context, _ testInput) (bool, error) {
	return true, nil
}

func (errorHandleSPI) Handle(_ context.Context, _ testInput) (string, error) {
	return "", fmt.Errorf("spi handle failed")
}

func TestTemplateFirstMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []testSPI{
		secondarySPI{},
		primarySPI{},
		refundSPI{},
	}, testInput{scene: "REFUND"}, First)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "secondary" {
		t.Fatalf("expected first implementation, got %q", results[0])
	}
}

func TestTemplateAllMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []testSPI{
		secondarySPI{},
		primarySPI{},
		refundSPI{},
	}, testInput{scene: "REFUND"}, All)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if results[0] != "secondary" || results[1] != "primary" || results[2] != "refund" {
		t.Fatalf("expected registration order to be preserved, got %#v", results)
	}
}

func TestTemplateFirstMatchedMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []testSPI{
		secondarySPI{},
		primarySPI{},
		refundSPI{},
	}, testInput{scene: "REFUND"}, FirstMatched)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0] != "refund" {
		t.Fatalf("expected first matched implementation, got %q", results[0])
	}
}

func TestTemplateAllMatchedMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []testSPI{
		secondarySPI{},
		primarySPI{},
		refundSPI{},
	}, testInput{scene: "ORDER"}, AllMatched)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0] != "secondary" || results[1] != "primary" {
		t.Fatalf("expected matched implementations in registration order, got %#v", results)
	}
}

func TestTemplateDefaultMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	results, err := template(context.Background(), []testSPI{
		secondarySPI{},
		primarySPI{},
	}, testInput{scene: "ORDER"}, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected default mode to return 1 result, got %d", len(results))
	}
	if results[0] != "secondary" {
		t.Fatalf("expected default mode to behave like First, got %q", results[0])
	}
}

func TestTemplateInvalidMode(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	if _, err := template(context.Background(), nil, testInput{}, Mode("UNKNOWN")); err == nil {
		t.Fatal("expected invalid mode to fail")
	}
}

func TestTemplateNilInvoke(t *testing.T) {
	template := NewTemplate[testSPI, testInput, string](nil, nil)

	if _, err := template(context.Background(), nil, testInput{}, First); err == nil {
		t.Fatal("expected nil invoke func to fail")
	}
}

func TestTemplateMatchedModeRequiresMatch(t *testing.T) {
	template := NewTemplate[testSPI, testInput, string](nil, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	if _, err := template(context.Background(), nil, testInput{}, FirstMatched); err == nil {
		t.Fatal("expected matched mode to require match func")
	}
}

func TestTemplateReturnsMatchError(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	if _, err := template(context.Background(), []testSPI{errorMatchSPI{}}, testInput{scene: "ORDER"}, FirstMatched); err == nil {
		t.Fatal("expected spi match error to be returned")
	}
}

func TestTemplateReturnsInvokeError(t *testing.T) {
	template := NewTemplate(func(ctx context.Context, impl testSPI, input testInput) (bool, error) {
		return impl.Match(ctx, input)
	}, func(ctx context.Context, impl testSPI, input testInput) (string, error) {
		return impl.Handle(ctx, input)
	})

	if _, err := template(context.Background(), []testSPI{errorHandleSPI{}}, testInput{scene: "ORDER"}, All); err == nil {
		t.Fatal("expected spi invoke error to be returned")
	}
}
