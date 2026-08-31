package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
)

// fakeDriver always fails on ping — lets us test Checker without a real DB
// (the package intentionally registers no drivers; users do it in main).
type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) { return fakeConn{}, nil }

type fakeConn struct{}

func (fakeConn) Prepare(query string) (driver.Stmt, error) { return nil, errors.New("unsupported") }
func (fakeConn) Close() error                              { return nil }
func (fakeConn) Begin() (driver.Tx, error)                 { return nil, errors.New("unsupported") }
func (fakeConn) Ping(ctx context.Context) error            { return errors.New("simulated ping failure") }

func init() {
	sql.Register("ginger-fake-fail", fakeDriver{})
}

func TestConnectUnknownDriverFails(t *testing.T) {
	if _, err := Connect(Config{Driver: "nonsense-driver", DSN: "x"}); err == nil {
		t.Fatal("unknown driver must fail")
	}
}

func TestConnectEmptyDriverFails(t *testing.T) {
	if _, err := Connect(Config{}); !errors.Is(err, ErrNoDriver) {
		t.Fatalf("expected ErrNoDriver, got %v", err)
	}
}

func TestCheckerFailsWhenDBIsDown(t *testing.T) {
	db, err := sql.Open("ginger-fake-fail", "")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() { _ = db.Close() }()

	c := NewChecker(db)
	if c.Name() != "database" {
		t.Fatalf("name = %q", c.Name())
	}
	if err := c.Check(context.Background()); err == nil {
		t.Fatal("check must fail when ping fails")
	}
}
