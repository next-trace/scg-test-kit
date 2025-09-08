// Package db contains thin adapters around scg-database contracts used by tests.
package db

import (
	db_contract "github.com/next-trace/scg-database/contract"
	"github.com/next-trace/scg-test-kit/contract"
)

// db is a thin wrapper exposing the Connection via the contract.DB interface.
// The actual ephemeral lifecycle (create/migrate/drop) must be performed in the calling service
// and the resulting Connection injected through testkit options.
// This keeps scg-test-kit contract-only and free of concrete details.
type db struct{ conn db_contract.Connection }

func (e *db) Conn() db_contract.Connection { return e.conn }

// NewFromConn wraps an existing scg-database connection into the contract.DB interface.
func NewFromConn(conn db_contract.Connection) contract.DB { //nolint:ireturn
	return &db{conn: conn}
}
