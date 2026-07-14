package migrations

import "embed"

// FS contains the SQL migration files used by the service at startup.
//
//go:embed *.sql
var FS embed.FS
