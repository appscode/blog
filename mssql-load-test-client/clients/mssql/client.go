package mssql

import "database/sql"

// Client wraps the sql.DB connection
type Client struct {
	*sql.DB
}
