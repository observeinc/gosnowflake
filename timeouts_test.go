package gosnowflake

import (
	"context"
	"net"
	"net/http"
	"net/url"
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

// TestPostRestfulQueryCancelAgainstSilentServer reproduces a stuck cancel request: the
// cancel needs a new connection, the TLS handshake for it never completes, and the
// request waits for that connection. With CleanupTimeout set, the caller returns at
// once and the abort itself ends within the timeout.
func TestPostRestfulQueryCancelAgainstSilentServer(t *testing.T) {
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
		FuncPost:      postRestful,
		TokenAccessor: getSimpleTokenAccessor(),
	}
	abortDone := make(chan struct{})
	sr.FuncCancelQuery = func(ctx context.Context, sr *snowflakeRestful, requestID UUID, timeout time.Duration) error {
		defer close(abortDone)
		return cancelQuery(ctx, sr, requestID, timeout)
	}

	start := time.Now()
	_, err = postRestfulQuery(context.Background(), sr, &url.Values{}, nil, nil, 0, NewUUID(), &Config{CleanupTimeout: 300 * time.Millisecond})
	assertErrIsE(t, err, context.DeadlineExceeded)
	assertTrueE(t, time.Since(start) < time.Second, "caller waited for the abort")
	select {
	case <-abortDone:
	case <-time.After(5 * time.Second):
		t.Fatal("abort request against a silent server did not end")
	}
}
