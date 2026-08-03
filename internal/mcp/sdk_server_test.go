package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connectSDKServer runs an SDKServer over an in-memory transport and returns a
// connected client session. Everything below drives the server the way a real
// client does, over the wire, rather than calling handlers directly: the point
// of slice 2 is that the SDK now owns the protocol, and only a real session
// exercises that.
func connectSDKServer(t *testing.T, registry *Registry) *mcp.ClientSession {
	t.Helper()

	server := NewSDKServer("productplan-test", "0.0.0", registry)
	clientTransport, serverTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.run(ctx, serverTransport)
	}()

	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0.0.0"}, nil)
	session, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(func() {
		_ = session.Close()
		<-done
	})
	return session
}

func testRegistry(t *testing.T) *Registry {
	t.Helper()
	r := NewRegistry()

	r.RegisterFunc(
		Tool{Name: "plain", Description: "plain text tool", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"summary":"ok"}`), nil
		},
	)
	r.RegisterFunc(
		Tool{
			Name:         "structured",
			Description:  "declares an output schema",
			InputSchema:  InputSchema{Type: "object"},
			OutputSchema: &OutputSchema{Type: "object"},
		},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{"summary":"ok","data":{}}`), nil
		},
	)
	r.RegisterFunc(
		Tool{Name: "echo_args", Description: "returns its arguments", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, args map[string]any) (json.RawMessage, error) {
			return json.Marshal(args)
		},
	)
	r.RegisterFunc(
		Tool{Name: "failing", Description: "always fails", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			return nil, errTestFailure
		},
	)
	r.RegisterFunc(
		Tool{Name: "panicking", Description: "panics", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			panic("boom")
		},
	)
	return r
}

type testErr string

func (e testErr) Error() string { return string(e) }

const errTestFailure = testErr("upstream exploded")

// The migration's whole purpose: the SDK negotiates the current revision, where
// the hand-rolled server was pinned to 2025-11-25 forever.
func TestSDKServerNegotiatesCurrentRevision(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	init := session.InitializeResult()
	if init == nil {
		t.Fatal("expected an initialize result")
	}
	if init.ProtocolVersion < "2026-07-28" {
		t.Errorf("negotiated %q, want at least 2026-07-28", init.ProtocolVersion)
	}
	if init.ServerInfo == nil || init.ServerInfo.Name != "productplan-test" {
		t.Errorf("unexpected server info: %+v", init.ServerInfo)
	}
}

func TestSDKServerListsEveryRegisteredTool(t *testing.T) {
	registry := testRegistry(t)
	session := connectSDKServer(t, registry)

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	if len(res.Tools) != registry.Count() {
		t.Errorf("listed %d tools, registry has %d", len(res.Tools), registry.Count())
	}

	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
	}
	for _, want := range []string{"plain", "structured", "echo_args", "failing", "panicking"} {
		if !got[want] {
			t.Errorf("tool %q missing from tools/list", want)
		}
	}
}

func TestSDKServerReturnsTextForToolsWithoutOutputSchema(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	res := callTool(t, session, "plain", nil)
	if res.IsError {
		t.Fatalf("unexpected error result: %s", contentText(t, res))
	}
	if got := contentText(t, res); got != `{"summary":"ok"}` {
		t.Errorf("unexpected text content: %s", got)
	}
	if res.StructuredContent != nil {
		t.Errorf("expected no structuredContent for a tool without an output schema, got %v", res.StructuredContent)
	}
}

func TestSDKServerReturnsStructuredContentWhenSchemaDeclared(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	res := callTool(t, session, "structured", nil)
	if res.IsError {
		t.Fatalf("unexpected error result: %s", contentText(t, res))
	}
	if res.StructuredContent == nil {
		t.Fatal("expected structuredContent for a tool that declares an output schema")
	}
	// The text payload is still sent alongside, which is what keeps existing
	// text-only clients working.
	if got := contentText(t, res); !strings.Contains(got, `"summary":"ok"`) {
		t.Errorf("expected the payload as text too, got %s", got)
	}
}

func TestSDKServerPassesArgumentsThrough(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	res := callTool(t, session, "echo_args", map[string]any{"id": "123", "limit": float64(5)})
	if res.IsError {
		t.Fatalf("unexpected error result: %s", contentText(t, res))
	}

	var echoed map[string]any
	if err := json.Unmarshal([]byte(contentText(t, res)), &echoed); err != nil {
		t.Fatalf("unmarshal echoed args: %v", err)
	}
	if echoed["id"] != "123" {
		t.Errorf("expected id to survive, got %v", echoed["id"])
	}
	if echoed["limit"] != float64(5) {
		t.Errorf("expected limit to survive, got %v", echoed["limit"])
	}
}

// A tool that fails must come back as an error RESULT, not a JSON-RPC protocol
// error. The distinction matters to a caller: a protocol error says the request
// was malformed, an error result says the tool ran and failed.
func TestSDKServerReportsToolFailureAsErrorResult(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "failing"})
	if err != nil {
		t.Fatalf("expected an error result, got a protocol error: %v", err)
	}
	if !res.IsError {
		t.Error("expected IsError to be set")
	}
	if got := contentText(t, res); !strings.HasPrefix(got, "Error: ") {
		t.Errorf("expected the existing \"Error: \" prefix, got %q", got)
	}
}

// HG-1. A panicking handler must not take the server down and must not surface
// as a silent success. Registry.Call owns the recovery, and this asserts the
// SDK adapter actually routes through it; calling Handler.Handle directly would
// bypass it and this test is what would catch that.
func TestSDKServerRecoversHandlerPanic(t *testing.T) {
	session := connectSDKServer(t, testRegistry(t))

	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "panicking"})
	if err != nil {
		t.Fatalf("panic escaped as a protocol error: %v", err)
	}
	if !res.IsError {
		t.Fatal("a panicking handler reported success; the panic was swallowed")
	}
	if got := contentText(t, res); !strings.Contains(got, "internal error in panicking") {
		t.Errorf("expected a structured internal-error message, got %q", got)
	}
	if got := contentText(t, res); strings.Contains(got, "boom") {
		t.Errorf("panic value leaked to the caller: %q", got)
	}

	// The session must still be usable afterwards.
	if after := callTool(t, session, "plain", nil); after.IsError {
		t.Error("server unusable after a handler panic")
	}
}

func callTool(t *testing.T, session *mcp.ClientSession, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("tools/call %s: %v", name, err)
	}
	return res
}

func contentText(t *testing.T, res *mcp.CallToolResult) string {
	t.Helper()
	if len(res.Content) == 0 {
		t.Fatal("result carried no content")
	}
	text, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	return text.Text
}

// Ported from the deleted integration_test.go, and strengthened. The old test
// fed 10 requests through the hand-rolled loop, which processed them one line at
// a time, so it never actually exercised concurrency and its unsynchronised
// counter was safe by accident. The SDK dispatches each request in its own
// goroutine, so this is the first test here that genuinely runs handlers in
// parallel; the counter is atomic because under the SDK it has to be.
func TestSDKServerHandlesConcurrentCalls(t *testing.T) {
	const calls = 10

	var invocations atomic.Int64
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "counter", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			invocations.Add(1)
			return json.RawMessage(`{"ok":true}`), nil
		},
	)

	session := connectSDKServer(t, registry)

	var wg sync.WaitGroup
	errs := make(chan error, calls)
	for range calls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "counter"})
			if err != nil {
				errs <- err
				return
			}
			if res.IsError {
				errs <- fmt.Errorf("unexpected error result")
			}
		}()
	}
	wg.Wait()
	close(errs)

	for err := range errs {
		t.Errorf("concurrent call failed: %v", err)
	}
	if got := invocations.Load(); got != calls {
		t.Errorf("handler ran %d times, want %d", got, calls)
	}
}

// Ported from the deleted integration_test.go. The old version asserted on a
// scanner error string from the bufio loop, which no longer exists; what
// matters now is that cancelling the context stops the server rather than
// leaking the goroutine.
func TestSDKServerRunStopsOnContextCancel(t *testing.T) {
	registry := NewRegistry()
	registry.RegisterFunc(
		Tool{Name: "noop", InputSchema: InputSchema{Type: "object"}},
		func(_ context.Context, _ map[string]any) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	)

	server := NewSDKServer("productplan-test", "0.0.0", registry)
	_, serverTransport := mcp.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = server.run(ctx, serverTransport)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return within 5s of the context being cancelled")
	}
}
