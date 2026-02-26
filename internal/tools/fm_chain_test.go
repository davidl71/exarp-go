package tools

import (
	"context"
	"fmt"
	"testing"
)

type mockGenerator struct {
	supported bool
	output    string
	err       error
}

func (m *mockGenerator) Supported() bool { return m.supported }
func (m *mockGenerator) Generate(_ context.Context, _ string, _ int, _ float32) (string, error) {
	return m.output, m.err
}

func TestChainFMProvider_FirstSuccess(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			&mockGenerator{supported: true, output: "first"},
			&mockGenerator{supported: true, output: "second"},
		},
	}

	out, err := chain.Generate(context.Background(), "test", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "first" {
		t.Errorf("got %q, want 'first'", out)
	}
}

func TestChainFMProvider_FallbackOnError(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			&mockGenerator{supported: true, err: fmt.Errorf("fail")},
			&mockGenerator{supported: true, output: "fallback"},
		},
	}

	out, err := chain.Generate(context.Background(), "test", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "fallback" {
		t.Errorf("got %q, want 'fallback'", out)
	}
}

func TestChainFMProvider_SkipUnsupported(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			&mockGenerator{supported: false, output: "skip"},
			&mockGenerator{supported: true, output: "used"},
		},
	}

	out, err := chain.Generate(context.Background(), "test", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "used" {
		t.Errorf("got %q, want 'used'", out)
	}
}

func TestChainFMProvider_AllFail(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			&mockGenerator{supported: true, err: fmt.Errorf("fail1")},
			&mockGenerator{supported: true, err: fmt.Errorf("fail2")},
		},
	}

	_, err := chain.Generate(context.Background(), "test", 10, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "fail2" {
		t.Errorf("got %q, want 'fail2' (last error)", err.Error())
	}
}

func TestChainFMProvider_NilBackends(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			nil,
			&mockGenerator{supported: true, output: "ok"},
		},
	}

	out, err := chain.Generate(context.Background(), "test", 10, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "ok" {
		t.Errorf("got %q, want 'ok'", out)
	}
}

func TestChainFMProvider_Empty(t *testing.T) {
	chain := &chainFMProvider{backends: []TextGenerator{}}

	_, err := chain.Generate(context.Background(), "test", 10, 0)
	if err != ErrFMNotSupported {
		t.Errorf("got %v, want ErrFMNotSupported", err)
	}
}

func TestChainFMProvider_Supported(t *testing.T) {
	chain := &chainFMProvider{
		backends: []TextGenerator{
			&mockGenerator{supported: false},
			&mockGenerator{supported: true},
		},
	}
	if !chain.Supported() {
		t.Error("chain should report supported when any backend supports")
	}
}

func TestChainStubFMProvider(t *testing.T) {
	stub := &chainStubFMProvider{}
	if stub.Supported() {
		t.Error("stub should not be supported")
	}

	_, err := stub.Generate(context.Background(), "test", 10, 0)
	if err != ErrFMNotSupported {
		t.Errorf("got %v, want ErrFMNotSupported", err)
	}
}
