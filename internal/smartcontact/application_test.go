package smartcontact

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// ---- fake SQL driver (mocks the DB behind database/sql's driver interfaces) ----

type fakeDriver struct{}
type fakeConn struct{}
type fakeStmt struct{}
type fakeRows struct{}

func (fakeDriver) Open(string) (driver.Conn, error)        { return fakeConn{}, nil }
func (fakeConn) Prepare(string) (driver.Stmt, error)        { return fakeStmt{}, nil }
func (fakeConn) Close() error                               { return nil }
func (fakeConn) Begin() (driver.Tx, error)                  { return nil, errors.New("unsupported") }
func (fakeStmt) Close() error                               { return nil }
func (fakeStmt) NumInput() int                              { return -1 }
func (fakeStmt) Exec([]driver.Value) (driver.Result, error) { return driver.RowsAffected(0), nil }
func (fakeStmt) Query([]driver.Value) (driver.Rows, error)  { return fakeRows{}, nil }
func (fakeRows) Columns() []string                          { return nil }
func (fakeRows) Close() error                               { return nil }
func (fakeRows) Next([]driver.Value) error                  { return io.EOF }

var registerOnce sync.Once

func fakeDB(t *testing.T) *sql.DB {
	t.Helper()
	registerOnce.Do(func() { sql.Register("smartcontact-fake", fakeDriver{}) })
	conn, err := sql.Open("smartcontact-fake", "")
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}
	return conn
}

func freeAddr(t *testing.T) (string, net.Listener) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	return l.Addr().String(), l
}

func init() { gin.SetMode(gin.TestMode) }

// ---- tests ----

func TestConstants(t *testing.T) {
	assert.Equal(t, "8080", defaultPort, "default port must match Spring Boot default")
	assert.Equal(t, 10*time.Second, shutdownTimeout)
}

// Context-load equivalent: wiring all components must not panic and must yield a handler.
func TestNewHandler_ContextLoads(t *testing.T) {
	conn := fakeDB(t)
	defer conn.Close()

	var h http.Handler
	assert.NotPanics(t, func() { h = NewHandler(conn) })
	assert.NotNil(t, h)
}

func TestApplicationRun(t *testing.T) {
	tests := []struct {
		name      string
		occupy    bool // port already in use
		wantErr   bool
		errSubstr string
	}{
		{name: "starts and shuts down gracefully on cancel", occupy: false, wantErr: false},
		{name: "fails when port already in use", occupy: true, wantErr: true, errSubstr: "http server"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			addr, l := freeAddr(t)
			if !tc.occupy {
				l.Close()
			} else {
				defer l.Close()
			}

			conn := fakeDB(t)
			app := &Application{
				db: conn,
				server: &http.Server{
					Addr:              addr,
					Handler:           NewHandler(conn),
					ReadHeaderTimeout: time.Second,
				},
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			done := make(chan error, 1)
			go func() { done <- app.Run(ctx) }()

			if !tc.occupy {
				var resp *http.Response
				var err error
				for i := 0; i < 50; i++ {
					resp, err = http.Get("http://" + addr + "/healthz")
					if err == nil {
						break
					}
					time.Sleep(20 * time.Millisecond)
				}
				if assert.NoError(t, err, "server should be listening") {
					body, _ := io.ReadAll(resp.Body)
					resp.Body.Close()
					assert.Equal(t, http.StatusOK, resp.StatusCode)
					assert.Equal(t, "ok", string(body))
				}
				cancel()
			}

			select {
			case err := <-done:
				if tc.wantErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tc.errSubstr)
				} else {
					assert.NoError(t, err)
				}
			case <-time.After(5 * time.Second):
				t.Fatal("Run did not return")
			}

			// Database must be released after Run returns.
			assert.Error(t, conn.Ping(), "db should be closed after Run")
		})
	}
}

// Error case: startup fails when configuration/datasource is invalid or unreachable.
// Depends on the environment's configuration, so failures are tolerated as errors, not panics.
func TestNewAndRun_StartupErrorsPropagate(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled ctx: ping fails or Run returns promptly

	tests := []struct {
		name string
		fn   func() error
	}{
		{"New", func() error {
			app, err := New(ctx)
			if err == nil && app != nil {
				_ = app.db.Close()
			}
			return err
		}},
		{"Run", func() error { return Run(ctx) }},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var err error
			assert.NotPanics(t, func() { err = tc.fn() })
			if err == nil {
				t.Skip("environment provides a working configuration; no startup error to assert")
			}
			assert.Error(t, err)
		})
	}
}