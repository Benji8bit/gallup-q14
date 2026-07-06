package migrations

import "embed"

// Files stores embedded SQL migrations.
//
//go:embed *.sql
var Files embed.FS
