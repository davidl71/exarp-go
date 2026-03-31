// embed.go — build-time embedded SQL migrations for exarp-go binaries (go install / PATH).
//
// When EXARP_MIGRATIONS_DIR is unset and no on-disk migrations directory is found next to the
// binary, EXARP_GO_ROOT, or project root, internal/database uses these files. See
// internal/database/migrations_resolve.go.
package migrations

import "embed"

// Files is the embedded migrations directory (NNN_description.sql).
//
//go:embed "*.sql"
var Files embed.FS
