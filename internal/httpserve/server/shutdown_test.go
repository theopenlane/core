package server

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	echo "github.com/theopenlane/echox"
	"gotest.tools/v3/assert"
	is "gotest.tools/v3/assert/cmp"

	coreconfig "github.com/theopenlane/core/v2/config"
	"github.com/theopenlane/core/v2/internal/httpserve/config"
)

// waitTimeout bounds every blocking assertion so a regression hangs the test instead of the suite
const waitTimeout = 5 * time.Second

// settleDelay is how long a negative "has not happened yet" check observes before concluding
const settleDelay = 100 * time.Millisecond

// errListener stands in for a non-shutdown failure returned by http.Server.Serve
var errListener = errors.New("listener failed")

// newShutdownServer builds a Server carrying only the shutdown grace period the drain helpers read
func newShutdownServer(grace time.Duration) *Server {
	return &Server{config: config.Config{Settings: coreconfig.Config{Server: coreconfig.Server{ShutdownGracePeriod: grace}}}}
}

// blockingHandler holds every request until released and reports when the first request has arrived
type blockingHandler struct {
	started chan struct{}
	release chan struct{}
}

// newBlockingHandler returns a handler whose requests block until release is closed
func newBlockingHandler() *blockingHandler {
	return &blockingHandler{started: make(chan struct{}), release: make(chan struct{})}
}

func (h *blockingHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	select {
	case <-h.started:
	default:
		close(h.started)
	}

	<-h.release

	w.WriteHeader(http.StatusOK)
}

// isClosed reports whether ch has been closed without blocking
func isClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
		return false
	}
}

// waitClosed fails the test unless ch closes within waitTimeout
func waitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(waitTimeout):
		t.Fatalf("timed out waiting for %s", what)
	}
}

// getInBackground issues a GET and delivers the status code once the response arrives
func getInBackground(t *testing.T, url string) <-chan int {
	t.Helper()

	status := make(chan int, 1)

	go func() {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
		if err != nil {
			status <- 0

			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			status <- 0

			return
		}

		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		status <- resp.StatusCode
	}()

	return status
}

func TestAwaitDrain(t *testing.T) {
	t.Parallel()

	t.Run("waits for the drain when the server was shut down", func(t *testing.T) {
		t.Parallel()

		drained := make(chan struct{})
		returned := make(chan error, 1)

		go func() { returned <- awaitDrain(http.ErrServerClosed, drained) }()

		select {
		case <-returned:
			t.Fatal("awaitDrain returned before the drain finished")
		case <-time.After(settleDelay):
		}

		close(drained)

		select {
		case err := <-returned:
			assert.Check(t, is.ErrorIs(err, http.ErrServerClosed))
		case <-time.After(waitTimeout):
			t.Fatal("awaitDrain did not return after the drain finished")
		}
	})

	t.Run("returns other errors without waiting", func(t *testing.T) {
		t.Parallel()

		returned := make(chan error, 1)

		go func() { returned <- awaitDrain(errListener, make(chan struct{})) }()

		select {
		case err := <-returned:
			assert.Check(t, is.ErrorIs(err, errListener))
		case <-time.After(waitTimeout):
			t.Fatal("awaitDrain blocked on a non-shutdown error")
		}
	})

	t.Run("returns nil without waiting", func(t *testing.T) {
		t.Parallel()

		returned := make(chan error, 1)

		go func() { returned <- awaitDrain(nil, make(chan struct{})) }()

		select {
		case err := <-returned:
			assert.Check(t, err)
		case <-time.After(waitTimeout):
			t.Fatal("awaitDrain blocked on a nil error")
		}
	})
}

func TestShutdownOnCancel(t *testing.T) {
	t.Parallel()

	t.Run("holds the drain open until in-flight requests finish", func(t *testing.T) {
		t.Parallel()

		handler := newBlockingHandler()
		srv, addr := listenAndServe(t, handler)

		ctx, cancel := context.WithCancel(context.Background())
		drained := make(chan struct{})

		assert.NilError(t, newShutdownServer(waitTimeout).shutdownOnCancel(ctx, drained)(srv.server))

		status := getInBackground(t, "http://"+addr)
		waitClosed(t, handler.started, "request to reach the handler")

		cancel()

		waitClosed(t, srv.served, "Serve to return after Shutdown closed the listener")
		assert.Check(t, is.ErrorIs(srv.err(), http.ErrServerClosed))

		time.Sleep(settleDelay)
		assert.Check(t, !isClosed(drained), "drain finished while a request was still in flight")

		close(handler.release)

		waitClosed(t, drained, "drain to finish after the handler returned")
		assert.Check(t, is.Equal(http.StatusOK, <-status))
	})

	t.Run("gives up on the drain when the grace period elapses", func(t *testing.T) {
		t.Parallel()

		handler := newBlockingHandler()
		defer close(handler.release)

		srv, addr := listenAndServe(t, handler)

		ctx, cancel := context.WithCancel(context.Background())
		drained := make(chan struct{})

		assert.NilError(t, newShutdownServer(settleDelay).shutdownOnCancel(ctx, drained)(srv.server))

		getInBackground(t, "http://"+addr)
		waitClosed(t, handler.started, "request to reach the handler")

		cancel()

		waitClosed(t, drained, "drain to give up after the grace period")
		assert.Check(t, !isClosed(handler.release), "handler was released before the drain gave up")
	})

	t.Run("shuts down immediately when the context is already cancelled", func(t *testing.T) {
		t.Parallel()

		srv, _ := listenAndServe(t, http.NotFoundHandler())

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		drained := make(chan struct{})

		assert.NilError(t, newShutdownServer(waitTimeout).shutdownOnCancel(ctx, drained)(srv.server))

		waitClosed(t, drained, "drain to finish with no requests in flight")
		waitClosed(t, srv.served, "Serve to return")
		assert.Check(t, is.ErrorIs(srv.err(), http.ErrServerClosed))
	})

	t.Run("does nothing until the context is cancelled", func(t *testing.T) {
		t.Parallel()

		srv, addr := listenAndServe(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		ctx, cancel := context.WithCancel(context.Background())
		drained := make(chan struct{})

		assert.NilError(t, newShutdownServer(waitTimeout).shutdownOnCancel(ctx, drained)(srv.server))

		assert.Check(t, is.Equal(http.StatusOK, <-getInBackground(t, "http://"+addr)))
		assert.Check(t, !isClosed(drained))
		assert.Check(t, !isClosed(srv.served))

		cancel()

		waitClosed(t, drained, "drain to finish")
	})
}

// TestStartEchoServerDrainOrdering runs the same StartConfig composition StartEchoServer uses through echox
// and proves Start does not return to the caller until in-flight requests have finished
func TestStartEchoServerDrainOrdering(t *testing.T) {
	t.Parallel()

	handler := newBlockingHandler()

	e := echo.New()
	e.Any("/*", echo.WrapHandler(handler))

	ctx, cancel := context.WithCancel(context.Background())
	drained := make(chan struct{})
	addr := make(chan net.Addr, 1)

	sc := echo.StartConfig{
		HideBanner:       true,
		HidePort:         true,
		Address:          "127.0.0.1:0",
		ListenerAddrFunc: func(a net.Addr) { addr <- a },
		BeforeServeFunc:  newShutdownServer(waitTimeout).shutdownOnCancel(ctx, drained),
	}

	returned := make(chan error, 1)

	go func() { returned <- awaitDrain(sc.Start(e), drained) }()

	var listenAddr net.Addr

	select {
	case listenAddr = <-addr:
	case <-time.After(waitTimeout):
		t.Fatal("server did not start listening")
	}

	status := getInBackground(t, "http://"+listenAddr.String())
	waitClosed(t, handler.started, "request to reach the handler")

	cancel()

	select {
	case <-returned:
		t.Fatal("Start returned while a request was still in flight")
	case <-time.After(settleDelay):
	}

	close(handler.release)

	select {
	case err := <-returned:
		assert.Check(t, is.ErrorIs(err, http.ErrServerClosed))
	case <-time.After(waitTimeout):
		t.Fatal("Start did not return after the drain finished")
	}

	assert.Check(t, is.Equal(http.StatusOK, <-status))
}

// servedServer is an http.Server serving on a loopback listener with its Serve result captured
type servedServer struct {
	server *http.Server
	served chan struct{}
	result error
}

// err returns the Serve result once served has closed
func (s *servedServer) err() error {
	<-s.served

	return s.result
}

// listenAndServe starts handler on a random loopback port and returns the server with its address
func listenAndServe(t *testing.T, handler http.Handler) (*servedServer, string) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NilError(t, err)

	s := &servedServer{
		server: &http.Server{Handler: handler, ReadHeaderTimeout: waitTimeout},
		served: make(chan struct{}),
	}

	go func() {
		defer close(s.served)

		s.result = s.server.Serve(listener)
	}()

	t.Cleanup(func() { _ = s.server.Close() })

	return s, listener.Addr().String()
}
