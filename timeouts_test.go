package gosnowflake

import (
	"context"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"testing"
	"time"
)

// startSilentServer accepts TCP connections and never writes to them, like a server
// that takes the TLS ClientHello and never sends a ServerHello.
func startSilentServer(t *testing.T) *net.TCPAddr {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assertNilF(t, err)
	var mu sync.Mutex
	var conns []net.Conn
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			mu.Lock()
			conns = append(conns, conn)
			mu.Unlock()
		}
	}()
	t.Cleanup(func() {
		_ = ln.Close()
		mu.Lock()
		defer mu.Unlock()
		for _, conn := range conns {
			_ = conn.Close()
		}
	})
	return ln.Addr().(*net.TCPAddr)
}

func TestTransportSetsTLSHandshakeTimeout(t *testing.T) {
	cfg := &Config{
		DisableOCSPChecks:       true,
		CertRevocationCheckMode: CertRevocationCheckDisabled,
	}
	for _, tt := range []transportType{transportTypeSnowflake, transportTypeCloudProvider} {
		rt, err := newTransportFactory(cfg, nil).createTransport(transportConfigFor(tt))
		assertNilF(t, err)
		transport, ok := rt.(*http.Transport)
		assertTrueF(t, ok, "expected *http.Transport")
		assertEqualE(t, transport.TLSHandshakeTimeout, http.DefaultTransport.(*http.Transport).TLSHandshakeTimeout)
	}
}

func TestTLSHandshakeTimeoutEndsSilentServer(t *testing.T) {
	addr := startSilentServer(t)
	tf := newTransportFactory(&Config{DisableOCSPChecks: true}, nil)
	transport, err := tf.createBaseTransport(&transportConfig{TLSHandshakeTimeout: 100 * time.Millisecond}, nil)
	assertNilF(t, err)
	client := &http.Client{Transport: transport}

	start := time.Now()
	_, err = client.Get("https://" + addr.String())
	assertNotNilF(t, err)
	assertStringContainsE(t, err.Error(), "TLS handshake timeout")
	assertTrueE(t, time.Since(start) < 5*time.Second, "handshake was not bounded")
}

func TestPostRestfulQueryBoundsCancelRequest(t *testing.T) {
	origTimeout := cancelQueryTimeout
	cancelQueryTimeout = 50 * time.Millisecond
	defer func() { cancelQueryTimeout = origTimeout }()

	var cancelCtxErr error
	var cancelCtxHasDeadline bool
	sr := &snowflakeRestful{
		FuncPostQueryHelper: func(context.Context, *snowflakeRestful, *url.Values, map[string]string, []byte, time.Duration, UUID, *Config) (*execResponse, error) {
			return nil, context.Canceled
		},
		FuncCancelQuery: func(ctx context.Context, _ *snowflakeRestful, _ UUID, _ time.Duration) error {
			cancelCtxErr = ctx.Err()
			_, cancelCtxHasDeadline = ctx.Deadline()
			// A cancel request that never gets an answer.
			<-ctx.Done()
			return ctx.Err()
		},
		TokenAccessor: getSimpleTokenAccessor(),
	}
	queryCtx, cancelQueryCtx := context.WithCancel(context.Background())
	cancelQueryCtx()

	start := time.Now()
	_, err := postRestfulQuery(queryCtx, sr, &url.Values{}, nil, nil, 0, NewUUID(), &Config{})
	assertNilE(t, cancelCtxErr, "cancel request must not inherit the query's done context")
	assertTrueE(t, cancelCtxHasDeadline, "cancel request must have a deadline")
	assertTrueE(t, time.Since(start) < 5*time.Second, "cancel request was not bounded")
	assertErrIsE(t, err, context.Canceled)
	assertErrIsE(t, err, context.DeadlineExceeded)
}

// TestPostRestfulQueryCancelAgainstSilentServer reproduces a stuck cancel request: the
// cancel needs a new connection, the TLS handshake for it never completes, and the
// request waits for that connection.
func TestPostRestfulQueryCancelAgainstSilentServer(t *testing.T) {
	origTimeout := cancelQueryTimeout
	cancelQueryTimeout = 300 * time.Millisecond
	defer func() { cancelQueryTimeout = origTimeout }()

	addr := startSilentServer(t)
	tf := newTransportFactory(&Config{DisableOCSPChecks: true}, nil)
	transport, err := tf.createBaseTransport(&transportConfig{}, nil)
	assertNilF(t, err)
	sr := &snowflakeRestful{
		Host:     addr.IP.String(),
		Port:     addr.Port,
		Protocol: "https",
		Client:   &http.Client{Transport: transport},
		FuncPostQueryHelper: func(context.Context, *snowflakeRestful, *url.Values, map[string]string, []byte, time.Duration, UUID, *Config) (*execResponse, error) {
			return nil, context.DeadlineExceeded
		},
		FuncPost:        postRestful,
		FuncCancelQuery: cancelQuery,
		TokenAccessor:   getSimpleTokenAccessor(),
	}

	start := time.Now()
	_, err = postRestfulQuery(context.Background(), sr, &url.Values{}, nil, nil, 0, NewUUID(), &Config{})
	elapsed := time.Since(start)
	assertNotNilF(t, err)
	assertErrIsE(t, err, context.DeadlineExceeded)
	assertTrueE(t, elapsed < 5*time.Second, "cancel request took "+strconv.FormatInt(elapsed.Milliseconds(), 10)+"ms")
}
