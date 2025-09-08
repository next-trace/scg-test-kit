// Package db provides a thin, contract-safe wrapper around scg-database.
//
// This package deliberately avoids owning any configuration or concrete
// database adapters/drivers. Microservices are responsible for constructing
// a db_contract.Connection (e.g., via their own helpers) and injecting it
// into the kit using testkit options. The wrapper here only exposes Conn().
package db
